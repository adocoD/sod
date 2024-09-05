package druid

import (
	"slices"
	"time"

	"github.com/wowsims/sod/sim/common/sod"
	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/stats"
)

// Totem Item IDs
const (
	WolfsheadHelm                    = 8345
	IdolOfFerocity                   = 22397
	IdolOfTheMoon                    = 23197
	IdolOfBrutality                  = 23198
	IdolMindExpandingMushroom        = 209576
	Catnip                           = 213407
	IdolOfWrath                      = 216490
	BloodBarkCrusher                 = 216499
	IdolOfTheDream                   = 220606
	RitualistsHammer                 = 221446
	Glasir                           = 224281
	Raelar                           = 224282
	IdolOfExsanguinationCat          = 228181
	IdolOfTheSwarm                   = 228180
	IdolOfExsanguinationBear         = 228182
	BloodGuardDragonhideGrips        = 227180
	KnightLieutenantsDragonhideGrips = 227183
	WushoolaysCharmOfNature          = 231280
	PristineEnchantedSouthSeasKelp   = 231316
)

func init() {
	core.AddEffectsToTest = false

	core.NewItemEffect(BloodBarkCrusher, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.newBloodbarkCleaveItem(BloodBarkCrusher)
	})

	// https://www.wowhead.com/classic/item=227180/blood-guards-dragonhide-grips
	// Equip: Reduces the mana cost of your shapeshifts by 150.
	core.NewItemEffect(BloodGuardDragonhideGrips, func(agent core.Agent) {
		registerDragonHideGripsAura(agent.(DruidAgent).GetDruid())
	})

	// https://www.wowhead.com/classic/item=224281/glasir
	// Equip: Critical effects from heals have a chance to heal 3 nearby allies for 200 to 350, and critical spell hits have a chance to damage 3 nearby enemies for 100 to 175 nature damage.
	// (Proc chance: 15%, 30s cooldown)
	core.NewItemEffect(Glasir, func(agent core.Agent) {
		character := agent.GetCharacter()

		numDamageHits := min(3, character.Env.GetNumTargets())
		numHealHits := min(3, len(character.Env.Raid.AllPlayerUnits))
		damageResults := make([]*core.SpellResult, numDamageHits)
		healResults := make([]*core.SpellResult, numHealHits)

		damageSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 452424},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				for idx := range damageResults {
					damageResults[idx] = spell.CalcDamage(sim, target, sim.Roll(100, 175), spell.OutcomeMagicHitAndCrit)
					target = sim.Environment.NextTargetUnit(target)
				}

				for _, result := range damageResults {
					spell.DealDamage(sim, result)
				}
			},
		})

		healSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 453009},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell | core.SpellFlagHelpful,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				for idx := range healResults {
					healResults[idx] = spell.CalcHealing(sim, target, sim.Roll(200, 350), spell.OutcomeMagicCrit)
					target = sim.Environment.NextTargetUnit(target)
				}

				for _, result := range healResults {
					spell.DealHealing(sim, result)
				}
			},
		})

		icd := core.Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Second * 30,
		}

		core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Gla'sir Damage Trigger",
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.ProcMask.Matches(core.ProcMaskSpellDamage) && result.DidCrit() && icd.IsReady(sim) && sim.Proc(.15, "Gla'sir Damage") {
					damageSpell.Cast(sim, result.Target)
					icd.Use(sim)
				}
			},
		}))

		core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Gla'sir Heal Trigger",
			OnHealDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if spell.ProcMask.Matches(core.ProcMaskSpellDamage) && result.DidCrit() && icd.IsReady(sim) && sim.Proc(.15, "Gla'sir Heal") {
					healSpell.Cast(sim, result.Target)
					icd.Use(sim)
				}
			},
		}))
	})

	// https://www.wowhead.com/classic/item=22397/idol-of-ferocity
	// Equip: Reduces the energy cost of Claw and Rake by 3.
	core.NewItemEffect(IdolOfFerocity, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()

		// TODO: Claw is not implemented
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidRake || spell.SpellCode == SpellCode_DruidMangleCat {
				spell.Cost.FlatModifier -= 3
			}
		})
	})

	// https://www.wowhead.com/classic/item=23197/idol-of-the-moon
	// Equip: Increases the damage of your Moonfire spell by up to 33.
	core.NewItemEffect(IdolOfTheMoon, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		affectedSpellCodes := []int32{SpellCode_DruidMoonfire, SpellCode_DruidSunfire, SpellCode_DruidStarfallSplash, SpellCode_DruidStarfallTick}
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if slices.Contains(affectedSpellCodes, spell.SpellCode) {
				spell.BonusDamage += 33
			}
		})
	})

	// https://www.wowhead.com/classic/item=23198/idol-of-brutality
	// Equip: Reduces the rage cost of Maul and Swipe by 3.
	core.NewItemEffect(IdolOfBrutality, func(agent core.Agent) {
		// Implemented in maul.go and swipe.go
	})

	core.NewItemEffect(IdolMindExpandingMushroom, func(agent core.Agent) {
		character := agent.GetCharacter()
		character.AddStat(stats.Spirit, 5)
	})

	// https://www.wowhead.com/classic/item=228181/idol-of-exsanguination-cat
	// Equip: The energy cost of your Rake and Rip spells is reduced by 5.
	core.NewItemEffect(IdolOfExsanguinationCat, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.OnSpellRegistered(func(spell *core.Spell) {
			if spell.SpellCode == SpellCode_DruidRake || spell.SpellCode == SpellCode_DruidRip {
				spell.Cost.FlatModifier -= 5
			}
		})
	})

	// https://www.wowhead.com/classic/item=228182/idol-of-exsanguination-bear
	// Equip: Your Lacerate ticks energize you for 3 rage.
	core.NewItemEffect(IdolOfExsanguinationBear, func(agent core.Agent) {
		// TODO: Not yet implemented
	})

	// https://www.wowhead.com/classic/item=228180/idol-of-the-swarm
	// Equip: The duration of your Insect Swarm spell is increased by 12 sec.
	core.NewItemEffect(IdolOfTheSwarm, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()

		bonusDuration := time.Second * 12

		druid.GetOrRegisterAura(core.Aura{
			Label: "Idol of the Swarm",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range druid.InsectSwarm {
					if spell != nil {
						for _, dot := range spell.Dots() {
							if dot != nil {
								dot.NumberOfTicks += 6
								dot.RecomputeAuraDuration()
							}
						}
					}
				}

				for _, aura := range druid.InsectSwarmAuras {
					if aura != nil && !aura.IsPermanent() {
						aura.Duration += bonusDuration
					}
				}
			},
		})
	})

	// https://www.wowhead.com/classic/item=227183/knight-lieutenants-dragonhide-grips
	// Equip: Reduces the mana cost of your shapeshifts by 150.
	core.NewItemEffect(KnightLieutenantsDragonhideGrips, func(agent core.Agent) {
		registerDragonHideGripsAura(agent.(DruidAgent).GetDruid())
	})

	// https://www.wowhead.com/classic/item=231316/pristine-enchanted-south-seas-kelp
	// Increases the critical hit chance of Wrath and Starfire by 2%.
	core.NewItemEffect(PristineEnchantedSouthSeasKelp, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.RegisterAura(core.Aura{
			Label: "Improved Wrath/Starfire",
			OnInit: func(aura *core.Aura, sim *core.Simulation) {
				for _, spell := range druid.Wrath {
					if spell != nil {
						spell.BonusCritRating += 2 * core.SpellCritRatingPerCritChance
					}
				}
				for _, spell := range druid.Starfire {
					if spell != nil {
						spell.BonusCritRating += 2 * core.SpellCritRatingPerCritChance
					}
				}
			},
		})
	})

	// https://www.wowhead.com/classic/item=224282/raelar
	// Equip: Chance on melee auto attack to steal 140 to 220 life from target enemy.
	core.NewItemEffect(Raelar, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{SpellID: 452430}
		healthMetrics := character.NewHealthMetrics(actionID)

		lifesteal := character.RegisterSpell(core.SpellConfig{
			ActionID:    actionID,
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				// TODO: Unsure if this can crit but it we're assuming Gla'sir can
				result := spell.CalcAndDealDamage(sim, target, sim.Roll(140, 220), spell.OutcomeMagicHitAndCrit)
				character.GainHealth(sim, result.Damage, healthMetrics)
			},
		})

		core.MakeProcTriggerAura(&character.Unit, core.ProcTrigger{
			Name:     "Rae'lar Damage Trigger",
			Callback: core.CallbackOnSpellHitDealt,
			ProcMask: core.ProcMaskWhiteHit,
			PPM:      1.0,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				lifesteal.Cast(sim, result.Target)
			},
		})
	})

	core.NewItemEffect(RitualistsHammer, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		druid.newBloodbarkCleaveItem(RitualistsHammer)
	})

	// https://www.wowhead.com/classic/item=231280/wushoolays-charm-of-nature
	// Use: Aligns the Druid with nature, increasing the damage done by spells by 15%, improving heal effects by 15%, and increasing the critical strike chance of spells by 10% for 20 sec.
	// (2 Min Cooldown)
	core.NewItemEffect(WushoolaysCharmOfNature, func(agent core.Agent) {
		character := agent.GetCharacter()
		actionID := core.ActionID{ItemID: WushoolaysCharmOfNature}
		duration := time.Second * 20

		aura := character.RegisterAura(core.Aura{
			ActionID: actionID,
			Label:    "Aligned with Nature",
			Duration: duration,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.SchoolDamageDealtMultiplier.MultiplyMagicSchools(1.15)
				// TODO: healing dealt multiplier?
				character.AddStatDynamic(sim, stats.SpellCrit, 10*core.SpellCritRatingPerCritChance)
			},
			OnExpire: func(aura *core.Aura, sim *core.Simulation) {
				character.PseudoStats.SchoolDamageDealtMultiplier.MultiplyMagicSchools(1 / 1.15)
				// TODO: healing dealt multiplier?
				character.AddStatDynamic(sim, stats.SpellCrit, -10*core.SpellCritRatingPerCritChance)
			},
		})

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    actionID,
			SpellSchool: core.SpellSchoolNature,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    character.NewTimer(),
					Duration: time.Minute * 2,
				},
				SharedCD: core.Cooldown{
					Timer:    character.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		character.AddMajorCooldown(core.MajorCooldown{
			Spell:    spell,
			Priority: core.CooldownPriorityBloodlust,
			Type:     core.CooldownTypeDPS,
		})
	})

	core.AddEffectsToTest = true
}

// https://www.wowhead.com/classic/item=213407/catnip
func (druid *Druid) registerCatnipCD() {
	if druid.Consumes.MiscConsumes == nil || !druid.Consumes.MiscConsumes.Catnip {
		return
	}
	sod.RegisterFiftyPercentHasteBuffCD(&druid.Character, core.ActionID{ItemID: Catnip})
}

func (druid *Druid) newBloodbarkCleaveItem(itemID int32) {
	auraActionID := core.ActionID{SpellID: 436482}

	results := make([]*core.SpellResult, min(3, druid.Env.GetNumTargets()))

	damageSpell := druid.RegisterSpell(Any, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 436481},
		SpellSchool: core.SpellSchoolPhysical | core.SpellSchoolNature,
		DefenseType: core.DefenseTypeMelee, // actually has DefenseTypeNone, but is likely using the greatest CritMultiplier available
		ProcMask:    core.ProcMaskEmpty,

		// TODO: "Causes additional threat" in Tooltip, no clue what the multiplier is.
		ThreatMultiplier: 1,
		DamageMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for idx := range results {
				results[idx] = spell.CalcDamage(sim, target, 5, spell.OutcomeMagicCrit)
				target = sim.Environment.NextTargetUnit(target)
			}
			for _, result := range results {
				spell.DealDamage(sim, result)
			}
		},
	})

	buffAura := druid.GetOrRegisterAura(core.Aura{
		Label:    "Bloodbark Cleave",
		ActionID: auraActionID,
		Duration: 20 * time.Second,

		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if result.Landed() && spell.ProcMask&core.ProcMaskMelee != 0 {
				damageSpell.Cast(sim, result.Target)
				return
			}
		},
	})

	mainSpell := druid.GetOrRegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{ItemID: itemID},
		Flags:    core.SpellFlagNoOnCastComplete | core.SpellFlagOffensiveEquipment,

		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    druid.NewTimer(),
				Duration: time.Minute * 3,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			buffAura.Activate(sim)
		},
	})

	druid.AddMajorCooldown(core.MajorCooldown{
		Spell:    mainSpell,
		Priority: core.CooldownPriorityDefault,
		Type:     core.CooldownTypeDPS,
	})
}

func registerDragonHideGripsAura(druid *Druid) {
	const costReduction int32 = 150
	var affectedForms []*DruidSpell

	druid.RegisterAura(core.Aura{
		Label:    "Dragonhide Grips",
		ActionID: core.ActionID{SpellID: 459594},
		Duration: core.NeverExpires,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedForms = []*DruidSpell{
				druid.CatForm,
				druid.MoonkinForm,
				druid.BearForm,
			}
		},
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedForms {
				if spell != nil {
					spell.Cost.FlatModifier -= costReduction
				}
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range affectedForms {
				if spell != nil {
					spell.Cost.FlatModifier += costReduction
				}
			}
		},
	})
}
