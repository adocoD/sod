import { Phase } from '../core/constants/other.js';
import * as PresetUtils from '../core/preset_utils.js';
import {
	AgilityElixir,
	AttackPowerBuff,
	Conjured,
	Consumes,
	Debuffs,
	EnchantedSigil,
	Explosive,
	FirePowerBuff,
	Flask,
	Food,
	IndividualBuffs,
	Potions,
	Profession,
	RaidBuffs,
	SaygesFortune,
	SpellPowerBuff,
	StrengthBuff,
	TristateEffect,
	WeaponImbue,
	ZanzaBuff,
} from '../core/proto/common.js';
import { PaladinAura, PaladinSeal, PaladinOptions as RetributionPaladinOptions } from '../core/proto/paladin.js';
import { SavedTalents } from '../core/proto/ui.js';
import APLP1RetJson from './apls/p1ret.apl.json';
import APLP2RetJson from './apls/p2ret.apl.json';
import APLP3RetJson from './apls/p3ret.apl.json';
import APLP4RetJson from './apls/p4ret.apl.json';
import APLP4RetExodinJson from './apls/p4ret-exodin.apl.json';
import APLP4RetExodin6PcT1Json from './apls/p4ret-exodin-6pcT1.apl.json';
import APLP4RetTwisting6PcT1Json from './apls/p4ret-twisting-6pcT1.apl.json';
import APLPP5ShockadinJson from './apls/p5Shockadin.apl.json';
import Phase1RetGearJson from './gear_sets/p1ret.gear.json';
import Phase2RetSoCGearJson from './gear_sets/p2retsoc.gear.json';
import Phase2RetSoMGearJson from './gear_sets/p2retsom.gear.json';
import Phase3RetSoMGearJson from './gear_sets/p3retsom.gear.json';
import Phase4RetExodinGearJson from './gear_sets/p4ret-exodin.gear.json';
import Phase4RetExodin6PcT1GearJson from './gear_sets/p4ret-exodin-6pcT1.gear.json';
import Phase4RetTwisting6PcT1GearJson from './gear_sets/p4ret-twisting-6pcT1.gear.json';
import Phase4RetGearJson from './gear_sets/p4rettwist.gear.json';
import Phase5ShockadinGearJson from './gear_sets/p5shockadin.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

///////////////////////////////////////////////////////////////////////////
//                                 Gear Presets
///////////////////////////////////////////////////////////////////////////

export const Phase1RetGear = PresetUtils.makePresetGear('P1', Phase1RetGearJson, {
	customCondition: player => player.getLevel() == 25,
});
export const Phase2RetSoCGear = PresetUtils.makePresetGear('P2 SoC/DS', Phase2RetSoCGearJson, {
	customCondition: player => player.getLevel() == 40,
});
export const Phase2RetSoMGear = PresetUtils.makePresetGear('P2 SoM', Phase2RetSoMGearJson, {
	customCondition: player => player.getLevel() == 40,
});
export const Phase3RetSoMGear = PresetUtils.makePresetGear('P3 SoM', Phase3RetSoMGearJson, {
	customCondition: player => player.getLevel() == 50,
});
export const Phase4RetTwistGear = PresetUtils.makePresetGear('P4 Twist', Phase4RetGearJson, {
	customCondition: player => player.getLevel() == 60,
});
export const Phase4RetTwist6pT1Gear = PresetUtils.makePresetGear('P4 Ret Twist 6pT1', Phase4RetTwisting6PcT1GearJson, {
	customCondition: player => player.getLevel() == 60,
});
export const Phase4RetExodinGear = PresetUtils.makePresetGear('P4 Ret Exodin', Phase4RetExodinGearJson, {
	customCondition: player => player.getLevel() == 60,
});
export const Phase4RetExodin6pT1Gear = PresetUtils.makePresetGear('P4 Ret Exodin 6pT1', Phase4RetExodin6PcT1GearJson, {
	customCondition: player => player.getLevel() == 60,
});
export const Phase5ShockadinGear = PresetUtils.makePresetGear('P5 Shockadin', Phase5ShockadinGearJson, {
	customCondition: player => player.getLevel() == 60,
});

export const GearPresets = {
	[Phase.Phase1]: [Phase1RetGear],
	[Phase.Phase2]: [Phase2RetSoCGear, Phase2RetSoMGear],
	[Phase.Phase3]: [Phase3RetSoMGear],
	[Phase.Phase4]: [Phase4RetTwistGear, Phase4RetTwist6pT1Gear, Phase4RetExodinGear, Phase4RetExodin6pT1Gear],
	[Phase.Phase5]: [Phase5ShockadinGear],
};

export const DefaultGear = GearPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 APL Presets
///////////////////////////////////////////////////////////////////////////

export const APLP1Ret = PresetUtils.makePresetAPLRotation('P1 Ret', APLP1RetJson, {
	customCondition: player => player.getLevel() === 25,
});
export const APLP2Ret = PresetUtils.makePresetAPLRotation('P2 Ret/Shockadin', APLP2RetJson, {
	customCondition: player => player.getLevel() === 40,
});
export const APLP3Ret = PresetUtils.makePresetAPLRotation('P3 Ret/Shockadin', APLP3RetJson, {
	customCondition: player => player.getLevel() === 50,
});
export const APLP4RetTwist = PresetUtils.makePresetAPLRotation('P4 Ret Twist', APLP4RetJson, {
	customCondition: player => player.getLevel() === 60,
});
export const APLP4RetTwist6pT1 = PresetUtils.makePresetAPLRotation('P4 Ret Twist 6pT1', APLP4RetTwisting6PcT1Json, {
	customCondition: player => player.getLevel() === 60,
});
export const APLP4RetExodin = PresetUtils.makePresetAPLRotation('P4 Ret Exodin', APLP4RetExodinJson, {
	customCondition: player => player.getLevel() === 60,
});
export const APLP4RetExodin6pT1 = PresetUtils.makePresetAPLRotation('P4 Ret Exodin 6pT1', APLP4RetExodin6PcT1Json, {
	customCondition: player => player.getLevel() === 60,
});
export const APLPP5Shockadin = PresetUtils.makePresetAPLRotation('P5 Shockadin', APLPP5ShockadinJson, {
	customCondition: player => player.getLevel() === 60,
});


export const APLPresets = {
	[Phase.Phase1]: [APLP1Ret],
	[Phase.Phase2]: [APLP2Ret],
	[Phase.Phase3]: [APLP3Ret],
	[Phase.Phase4]: [APLP4RetTwist, APLP4RetTwist6pT1, APLP4RetExodin, APLP4RetExodin6pT1],
	[Phase.Phase5]: [APLPP5Shockadin],
};

export const DefaultAPLs: Record<number, PresetUtils.PresetRotation> = {
	25: APLPresets[Phase.Phase1][0],
	40: APLPresets[Phase.Phase2][0],
	50: APLPresets[Phase.Phase3][0],
	60: APLPresets[Phase.Phase4][0],
};

///////////////////////////////////////////////////////////////////////////
//                                 Talent presets
///////////////////////////////////////////////////////////////////////////

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/classic/talent-calc and copy the numbers in the url.

export const P1RetTalents = PresetUtils.makePresetTalents('P1 Ret', SavedTalents.create({ talentsString: '--05230051' }), {
	customCondition: player => player.getLevel() === 25,
});

export const P2RetTalents = PresetUtils.makePresetTalents('P2 Ret', SavedTalents.create({ talentsString: '--532300512003151' }), {
	customCondition: player => player.getLevel() === 40,
});

export const P2ShockadinTalents = PresetUtils.makePresetTalents('P2 Shockadin', SavedTalents.create({ talentsString: '55050100521151--' }), {
	customCondition: player => player.getLevel() === 40,
});

export const P3RetTalents = PresetUtils.makePresetTalents('P3 Ret', SavedTalents.create({ talentsString: '500501--53230051200315' }), {
	customCondition: player => player.getLevel() === 50,
});

export const P4RetTalents = PresetUtils.makePresetTalents('P4 Ret', SavedTalents.create({ talentsString: '500501-503-52230351200315' }), {
	customCondition: player => player.getLevel() === 60,
});

export const P5ShockadinTalents = PresetUtils.makePresetTalents('P5 Shockadin', SavedTalents.create({ talentsString: '55053100501051--052303511' }), {
	customCondition: player => player.getLevel() === 60,
});


export const TalentPresets = {
	[Phase.Phase1]: [P1RetTalents],
	[Phase.Phase2]: [P2RetTalents, P2ShockadinTalents],
	[Phase.Phase3]: [P3RetTalents],
	[Phase.Phase4]: [P4RetTalents],
	[Phase.Phase5]: [P5ShockadinTalents],
};

// TODO: Phase 3
export const DefaultTalents = TalentPresets[Phase.Phase4][0];

///////////////////////////////////////////////////////////////////////////
//                                 Options
///////////////////////////////////////////////////////////////////////////

export const DefaultOptions = RetributionPaladinOptions.create({
	aura: PaladinAura.SanctityAura,
	primarySeal: PaladinSeal.Martyrdom,
});

export const DefaultConsumes = Consumes.create({
	agilityElixir: AgilityElixir.ElixirOfTheMongoose,
	boglingRoot: false,
	defaultPotion: Potions.MajorManaPotion,
	dragonBreathChili: true,
	enchantedSigil: EnchantedSigil.FlowingWatersSigil,
	food: Food.FoodBlessSunfruit,
	flask: Flask.FlaskOfSupremePower,
	firePowerBuff: FirePowerBuff.ElixirOfGreaterFirepower,
	fillerExplosive: Explosive.ExplosiveUnknown,
	mainHandImbue: WeaponImbue.WildStrikes,
	miscConsumes: {
		jujuEmber: true,
	},
	spellPowerBuff: SpellPowerBuff.GreaterArcaneElixir,
	strengthBuff: StrengthBuff.JujuPower,
	zanzaBuff: ZanzaBuff.ROIDS,
	attackPowerBuff: AttackPowerBuff.JujuMight,
	defaultConjured: Conjured.ConjuredDemonicRune,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfKings: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	fengusFerocity: true,
	mightOfStormwind: true,
	moldarsMoxie: true,
	rallyingCryOfTheDragonslayer: true,
	saygesFortune: SaygesFortune.SaygesDamage,
	slipkiksSavvy: true,
	songflowerSerenade: true,
	valorOfAzeroth: true,
	warchiefsBlessing: true,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	battleShout: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	fireResistanceAura: true,
	fireResistanceTotem: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	sanctityAura: true,
	leaderOfThePack: true,
	demonicPact: 80,
	aspectOfTheLion: true,
	moonkinAura: true,
	vampiricTouch: 300,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfRecklessness: true,
	homunculi: 70, // 70% average uptime default
	faerieFire: true,
	giftOfArthas: true,
	sunderArmor: true,
	judgementOfWisdom: true,
	judgementOfTheCrusader: TristateEffect.TristateEffectImproved,
	improvedFaerieFire: true,
	improvedScorch: true,
	markOfChaos: true,
	occultPoison: true,
	mangle: true,
});

export const OtherDefaults = {
	profession1: Profession.Blacksmithing,
	profession2: Profession.Enchanting,
};
