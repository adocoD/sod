import { default as pako } from 'pako';
import { ref } from 'tsx-vanilla';

import * as Mechanics from '../constants/mechanics';
import { SimSettingCategories } from '../constants/sim-settings';
import { IndividualSimUI } from '../individual_sim_ui';
import { RaidSimRequest } from '../proto/api';
import { PseudoStat, Spec, Stat } from '../proto/common';
import { IndividualSimSettings } from '../proto/ui';
import { classNames, raceNames } from '../proto_utils/names';
import { UnitStat } from '../proto_utils/stats';
import { specNames } from '../proto_utils/utils';
import { SimUI } from '../sim_ui';
import { EventID, TypedEvent } from '../typed_event';
import { arrayEquals, downloadString, getEnumValues, jsonStringifyWithFlattenedPaths } from '../utils';
import { BaseModal } from './base_modal';
import { BooleanPicker } from './boolean_picker';
import { CopyButton } from './copy_button';
import { IndividualLinkImporter, IndividualWowheadGearPlannerImporter } from './importers';

interface ExporterOptions {
	title: string;
	header?: boolean;
	allowDownload?: boolean;
}

export abstract class Exporter extends BaseModal {
	private readonly textElem: HTMLElement;
	protected readonly changedEvent: TypedEvent<void> = new TypedEvent();

	constructor(parent: HTMLElement, simUI: SimUI, options: ExporterOptions) {
		super(parent, 'exporter', { title: options.title, header: true, footer: true });

		this.body.innerHTML = `
			<textarea spellCheck="false" class="exporter-textarea form-control"></textarea>
		`;
		this.textElem = this.rootElem.getElementsByClassName('exporter-textarea')[0] as HTMLElement;

		new CopyButton(this.footer!, {
			extraCssClasses: ['btn-primary', 'me-2'],
			getContent: () => this.textElem.innerHTML,
			text: 'Copy',
			tooltip: 'Copy to clipboard',
		});

		if (options.allowDownload) {
			const downloadBtnRef = ref<HTMLButtonElement>();
			this.footer!.appendChild(
				<button className="exporter-button btn btn-primary download-button" ref={downloadBtnRef}>
					<i className="fa fa-download me-1"></i>
					Download
				</button>,
			);

			const downloadButton = downloadBtnRef.value!;
			downloadButton.addEventListener('click', _event => {
				const data = this.textElem.textContent!;
				downloadString(data, 'wowsims.json');
			});
		}
	}

	protected init() {
		this.changedEvent.on(() => this.updateContent());
		this.updateContent();
	}

	private updateContent() {
		this.textElem.textContent = this.getData();
	}

	abstract getData(): string;
}

export class IndividualLinkExporter<SpecType extends Spec> extends Exporter {
	private static readonly exportPickerConfigs: Array<{
		category: SimSettingCategories;
		label: string;
		labelTooltip: string;
	}> = [
		{
			category: SimSettingCategories.Gear,
			label: 'Gear',
			labelTooltip: 'Also includes bonus stats and weapon swaps.',
		},
		{
			category: SimSettingCategories.Talents,
			label: 'Talents',
			labelTooltip: 'Talents and Glyphs.',
		},
		{
			category: SimSettingCategories.Rotation,
			label: 'Rotation',
			labelTooltip: 'Includes everything found in the Rotation tab.',
		},
		{
			category: SimSettingCategories.Consumes,
			label: 'Consumes',
			labelTooltip: 'Flask, pots, food, etc.',
		},
		{
			category: SimSettingCategories.External,
			label: 'Buffs & Debuffs',
			labelTooltip: 'All settings which are applied by other raid members.',
		},
		{
			category: SimSettingCategories.Miscellaneous,
			label: 'Misc',
			labelTooltip: 'Spec-specific settings, front/back of target, distance from target, etc.',
		},
		{
			category: SimSettingCategories.Encounter,
			label: 'Encounter',
			labelTooltip: 'Fight-related settings.',
		},
		// Intentionally exclude UISettings category here, because users almost
		// never intend to export them and it messes with other users' settings.
		// If they REALLY want to export UISettings, just use the JSON exporter.
	];

	private readonly simUI: IndividualSimUI<SpecType>;
	private readonly exportCategories: Record<SimSettingCategories, boolean>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'Sharable Link' });
		this.simUI = simUI;

		const exportCategories: Partial<Record<SimSettingCategories, boolean>> = {};
		(getEnumValues(SimSettingCategories) as Array<SimSettingCategories>).forEach(
			cat => (exportCategories[cat] = IndividualLinkImporter.DEFAULT_CATEGORIES.includes(cat)),
		);
		this.exportCategories = exportCategories as Record<SimSettingCategories, boolean>;

		const pickersContainer = document.createElement('div');
		pickersContainer.classList.add('link-exporter-pickers');
		this.body.prepend(pickersContainer);

		IndividualLinkExporter.exportPickerConfigs.forEach(exportConfig => {
			const category = exportConfig.category;
			new BooleanPicker(pickersContainer, this, {
				id: `link-exporter-${category}`,
				label: exportConfig.label,
				labelTooltip: exportConfig.labelTooltip,
				inline: true,
				getValue: () => this.exportCategories[category],
				setValue: (eventID: EventID, _modObj: IndividualLinkExporter<SpecType>, newValue: boolean) => {
					this.exportCategories[category] = newValue;
					this.changedEvent.emit(eventID);
				},
				changedEvent: () => this.changedEvent,
			});
		});
	}

	open() {
		super.open();
		this.init();
	}

	getData(): string {
		return IndividualLinkExporter.createLink(
			this.simUI,
			(getEnumValues(SimSettingCategories) as Array<SimSettingCategories>).filter(c => this.exportCategories[c]),
		);
	}

	static createLink(simUI: IndividualSimUI<any>, exportCategories?: Array<SimSettingCategories>): string {
		if (!exportCategories) {
			exportCategories = IndividualLinkImporter.DEFAULT_CATEGORIES;
		}

		const proto = simUI.toProto(exportCategories);

		const protoBytes = IndividualSimSettings.toBinary(proto);
		// @ts-ignore Pako did some weird stuff between versions and the @types package doesn't correctly support this syntax for version 2.0.4 but it's completely valid
		// The syntax was removed in 2.1.0 and there were several complaints but the project seems to be largely abandoned now
		const deflated = pako.deflate(protoBytes, { to: 'string' });
		const encoded = btoa(String.fromCharCode(...deflated));

		const linkUrl = new URL(window.location.href);
		linkUrl.hash = encoded;
		if (arrayEquals(exportCategories, IndividualLinkImporter.DEFAULT_CATEGORIES)) {
			linkUrl.searchParams.delete(IndividualLinkImporter.CATEGORY_PARAM);
		} else {
			const categoryCharString = exportCategories.map(c => IndividualLinkImporter.CATEGORY_KEYS.get(c)).join('');
			linkUrl.searchParams.set(IndividualLinkImporter.CATEGORY_PARAM, categoryCharString);
		}
		return linkUrl.toString();
	}
}

export class IndividualJsonExporter<SpecType extends Spec> extends Exporter {
	private readonly simUI: IndividualSimUI<SpecType>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'JSON Export', allowDownload: true });
		this.simUI = simUI;
	}

	open() {
		super.open();
		this.init();
	}

	getData(): string {
		return jsonStringifyWithFlattenedPaths(IndividualSimSettings.toJson(this.simUI.toProto()), 2, (value, path) => {
			if (['stats', 'pseudoStats'].includes(path[path.length - 1])) {
				return true;
			}

			if (['player', 'equipment', 'items'].every((v, i) => path[i] == v)) {
				return path.length > 3;
			}

			if (path[0] == 'player' && path[1] == 'rotation' && ['prepullActions', 'priorityList'].includes(path[2])) {
				return path.length > 3;
			}

			return false;
		});
	}
}

export class IndividualWowheadGearPlannerExporter<SpecType extends Spec> extends Exporter {
	private readonly simUI: IndividualSimUI<SpecType>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'Wowhead Export', allowDownload: true });
		this.simUI = simUI;
	}

	open() {
		super.open();
		this.init();
	}

	getData(): string {
		const player = this.simUI.player;

		const classStr = classNames.get(player.getClass())!.replaceAll(' ', '-').toLowerCase();
		const raceStr = raceNames.get(player.getRace())!.replaceAll(' ', '-').toLowerCase();
		const url = `https://www.wowhead.com/classic/gear-planner/${classStr}/${raceStr}/`;

		// See comments on the importer for how the binary formatting is structured.
		let bytes: Array<number> = [];
		bytes.push(6);
		bytes.push(0);
		bytes.push(Mechanics.MAX_CHARACTER_LEVEL);

		let talentsStr = player.getTalentsString().replaceAll('-', 'f') + 'f';
		if (talentsStr.length % 2 == 1) {
			talentsStr += '0';
		}
		//console.log('Talents str: ' + talentsStr);
		bytes.push(talentsStr.length / 2);
		for (let i = 0; i < talentsStr.length; i += 2) {
			bytes.push(parseInt(talentsStr.substring(i, i + 2), 16));
		}

		const to2Bytes = (val: number): Array<number> => {
			return [(val & 0xff00) >> 8, val & 0x00ff];
		};

		const gear = player.getGear();
		gear.getItemSlots()
			.sort((slot1, slot2) => IndividualWowheadGearPlannerImporter.slotIDs[slot1] - IndividualWowheadGearPlannerImporter.slotIDs[slot2])
			.forEach(itemSlot => {
				const item = gear.getEquippedItem(itemSlot);
				if (!item) {
					return;
				}

				let slotId = IndividualWowheadGearPlannerImporter.slotIDs[itemSlot];
				if (item.enchant) {
					slotId = slotId | 0b10000000;
				}
				bytes.push(slotId);
				bytes.push(0 << 5);
				bytes = bytes.concat(to2Bytes(item.item.id));

				if (item.enchant) {
					bytes.push(0);
					bytes = bytes.concat(to2Bytes(item.enchant.spellId));
				}
			});

		//console.log('Hex: ' + buf2hex(new Uint8Array(bytes)));
		const binaryString = String.fromCharCode(...bytes);
		const b64encoded = btoa(binaryString);
		const b64converted = b64encoded.replaceAll('/', '_').replaceAll('+', '-').replace(/=+$/, '');

		return url + b64converted;
	}
}

export class Individual60UEPExporter<SpecType extends Spec> extends Exporter {
	private readonly simUI: IndividualSimUI<SpecType>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: '80Upgrades EP Export', allowDownload: true });
		this.simUI = simUI;
	}

	open() {
		super.open();
		this.init();
	}

	getData(): string {
		const player = this.simUI.player;
		const epValues = player.getEpWeights();
		const allUnitStats = UnitStat.getAll();

		const namesToWeights: Record<string, number> = {};
		allUnitStats.forEach(stat => {
			const statName = Individual60UEPExporter.getName(stat);
			const weight = epValues.getUnitStat(stat);
			if (weight == 0 || statName == '') {
				return;
			}

			// Need to add together stats with the same name (e.g. hit/crit/haste).
			if (namesToWeights[statName]) {
				namesToWeights[statName] += weight;
			} else {
				namesToWeights[statName] = weight;
			}
		});

		return (
			`https://sixtyupgrades.com/sod/ep/import?name=${encodeURIComponent(`${specNames[player.spec]} WoWSims Weights`)}` +
			Object.keys(namesToWeights)
				.map(statName => `&${statName}=${namesToWeights[statName].toFixed(3)}`)
				.join('')
		);
	}

	static getName(stat: UnitStat): string {
		if (stat.isStat()) {
			return Individual60UEPExporter.statNames[stat.getStat()];
		} else {
			return Individual60UEPExporter.pseudoStatNames[stat.getPseudoStat()] || '';
		}
	}

	static statNames: Record<Stat, string> = {
		[Stat.StatStrength]: 'strength',
		[Stat.StatAgility]: 'agility',
		[Stat.StatStamina]: 'stamina',
		[Stat.StatIntellect]: 'intellect',
		[Stat.StatSpirit]: 'spirit',
		[Stat.StatSpellPower]: 'spellPower',
		[Stat.StatSpellDamage]: 'spellDamage',
		[Stat.StatArcanePower]: 'arcaneDamage',
		[Stat.StatHolyPower]: 'holyDamage',
		[Stat.StatFirePower]: 'fireDamage',
		[Stat.StatFrostPower]: 'frostDamage',
		[Stat.StatNaturePower]: 'natureDamage',
		[Stat.StatShadowPower]: 'shadowDamage',
		[Stat.StatMP5]: 'mp5',
		[Stat.StatSpellHit]: 'spellHit',
		[Stat.StatSpellCrit]: 'spellCrit',
		[Stat.StatSpellHaste]: 'spellHaste',
		[Stat.StatSpellPenetration]: 'spellPen',
		[Stat.StatAttackPower]: 'attackPower',
		[Stat.StatMeleeHit]: 'hit',
		[Stat.StatMeleeCrit]: 'crit',
		[Stat.StatMeleeHaste]: 'haste',
		[Stat.StatArmorPenetration]: 'armorPen',
		[Stat.StatExpertise]: 'expertise',
		[Stat.StatMana]: 'mana',
		[Stat.StatEnergy]: 'energy',
		[Stat.StatRage]: 'rage',
		[Stat.StatArmor]: 'armor',
		[Stat.StatRangedAttackPower]: 'attackPower',
		[Stat.StatDefense]: 'defense',
		[Stat.StatBlock]: 'block',
		[Stat.StatBlockValue]: 'blockValue',
		[Stat.StatDodge]: 'dodge',
		[Stat.StatParry]: 'parry',
		[Stat.StatResilience]: 'resilience',
		[Stat.StatHealth]: 'health',
		[Stat.StatArcaneResistance]: 'arcaneResistance',
		[Stat.StatFireResistance]: 'fireResistance',
		[Stat.StatFrostResistance]: 'frostResistance',
		[Stat.StatNatureResistance]: 'natureResistance',
		[Stat.StatShadowResistance]: 'shadowResistance',
		[Stat.StatBonusArmor]: 'armorBonus',
		[Stat.StatHealingPower]: 'healing',
		[Stat.StatFeralAttackPower]: 'feralAttackPower',
	};
	static pseudoStatNames: Partial<Record<PseudoStat, string>> = {
		[PseudoStat.PseudoStatMainHandDps]: 'dps',
		[PseudoStat.PseudoStatRangedDps]: 'rangedDps',
		// Weapon Skills
		[PseudoStat.PseudoStatUnarmedSkill]: 'unarmedSkill',
		[PseudoStat.PseudoStatDaggersSkill]: 'daggerSkill',
		[PseudoStat.PseudoStatSwordsSkill]: 'swordSkill',
		[PseudoStat.PseudoStatMacesSkill]: 'maceSkill',
		[PseudoStat.PseudoStatAxesSkill]: 'axeSkill',
		[PseudoStat.PseudoStatTwoHandedSwordsSkill]: 'sword2hSkill',
		[PseudoStat.PseudoStatTwoHandedMacesSkill]: 'mace2hSkill',
		[PseudoStat.PseudoStatTwoHandedAxesSkill]: 'axe2hSkill',
		[PseudoStat.PseudoStatPolearmsSkill]: 'polearmSkill',
		[PseudoStat.PseudoStatStavesSkill]: 'staffSkill',
		[PseudoStat.PseudoStatBowsSkill]: 'bowSkill',
		[PseudoStat.PseudoStatCrossbowsSkill]: 'crossbowSkill',
		[PseudoStat.PseudoStatGunsSkill]: 'gunSkill',
		[PseudoStat.PseudoStatThrownSkill]: 'thrownSkill',
	};
}

export class IndividualPawnEPExporter<SpecType extends Spec> extends Exporter {
	private readonly simUI: IndividualSimUI<SpecType>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'Pawn EP Export', allowDownload: true });
		this.simUI = simUI;
	}

	open() {
		super.open();
		this.init();
	}

	getData(): string {
		const player = this.simUI.player;
		const epValues = player.getEpWeights();
		const allUnitStats = UnitStat.getAll();

		const namesToWeights: Record<string, number> = {};
		allUnitStats.forEach(stat => {
			const statName = IndividualPawnEPExporter.getName(stat);
			const weight = epValues.getUnitStat(stat);
			if (weight == 0 || statName == '') {
				return;
			}

			// Need to add together stats with the same name (e.g. hit/crit/haste).
			if (namesToWeights[statName]) {
				namesToWeights[statName] += weight;
			} else {
				namesToWeights[statName] = weight;
			}
		});

		return (
			`( Pawn: v1: "${specNames[player.spec]} WoWSims Weights": Class=${classNames.get(player.getClass())},` +
			Object.keys(namesToWeights)
				.map(statName => `${statName}=${namesToWeights[statName].toFixed(3)}`)
				.join(',') +
			' )'
		);
	}

	static getName(stat: UnitStat): string {
		if (stat.isStat()) {
			return IndividualPawnEPExporter.statNames[stat.getStat()];
		} else {
			return IndividualPawnEPExporter.pseudoStatNames[stat.getPseudoStat()] || '';
		}
	}

	static statNames: Record<Stat, string> = {
		[Stat.StatStrength]: 'Strength',
		[Stat.StatAgility]: 'Agility',
		[Stat.StatStamina]: 'Stamina',
		[Stat.StatIntellect]: 'Intellect',
		[Stat.StatSpirit]: 'Spirit',
		[Stat.StatSpellPower]: 'SpellDamage',
		[Stat.StatSpellDamage]: 'SpellDamage',
		[Stat.StatArcanePower]: 'ArcaneSpellDamage',
		[Stat.StatFirePower]: 'FireSpellDamage',
		[Stat.StatFrostPower]: 'FrostSpellDamage',
		[Stat.StatHolyPower]: 'HolySpellDamage',
		[Stat.StatNaturePower]: 'NatureSpellDamage',
		[Stat.StatShadowPower]: 'ShadowSpellDamage',
		[Stat.StatMP5]: 'Mp5',
		[Stat.StatSpellHit]: 'SpellHitRating',
		[Stat.StatSpellCrit]: 'SpellCritRating',
		[Stat.StatSpellHaste]: 'SpellHasteRating',
		[Stat.StatSpellPenetration]: 'SpellPen',
		[Stat.StatAttackPower]: 'Ap',
		[Stat.StatMeleeHit]: 'HitRating',
		[Stat.StatMeleeCrit]: 'CritRating',
		[Stat.StatMeleeHaste]: 'HasteRating',
		[Stat.StatArmorPenetration]: 'ArmorPenetration',
		[Stat.StatExpertise]: 'ExpertiseRating',
		[Stat.StatMana]: 'Mana',
		[Stat.StatEnergy]: 'Energy',
		[Stat.StatRage]: 'Rage',
		[Stat.StatArmor]: 'Armor',
		[Stat.StatRangedAttackPower]: 'Ap',
		[Stat.StatDefense]: 'DefenseRating',
		[Stat.StatBlock]: 'BlockRating',
		[Stat.StatBlockValue]: 'BlockValue',
		[Stat.StatDodge]: 'DodgeRating',
		[Stat.StatParry]: 'ParryRating',
		[Stat.StatResilience]: 'ResilienceRating',
		[Stat.StatHealth]: 'Health',
		[Stat.StatArcaneResistance]: 'ArcaneResistance',
		[Stat.StatFireResistance]: 'FireResistance',
		[Stat.StatFrostResistance]: 'FrostResistance',
		[Stat.StatNatureResistance]: 'NatureResistance',
		[Stat.StatShadowResistance]: 'ShadowResistance',
		[Stat.StatBonusArmor]: 'Armor2',
		[Stat.StatHealingPower]: 'Healing',
		[Stat.StatFeralAttackPower]: 'FeralAttackPower',
	};
	static pseudoStatNames: Partial<Record<PseudoStat, string>> = {
		[PseudoStat.PseudoStatMainHandDps]: 'MeleeDps',
		[PseudoStat.PseudoStatRangedDps]: 'RangedDps',
	};
}

export class IndividualCLIExporter<SpecType extends Spec> extends Exporter {
	private readonly simUI: IndividualSimUI<SpecType>;

	constructor(parent: HTMLElement, simUI: IndividualSimUI<SpecType>) {
		super(parent, simUI, { title: 'CLI Export', allowDownload: true });
		this.simUI = simUI;
	}

	open() {
		super.open();
		this.init();
	}

	getData(): string {
		const raidSimJson: any = RaidSimRequest.toJson(this.simUI.sim.makeRaidSimRequest(false));
		delete raidSimJson.raid?.parties[0]?.players[0]?.database;
		return JSON.stringify(raidSimJson, null, 2);
	}
}
