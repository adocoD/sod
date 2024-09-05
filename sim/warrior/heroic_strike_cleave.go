package warrior

import (
	"github.com/wowsims/sod/sim/core"
)

func (warrior *Warrior) registerHeroicStrikeSpell() {
	flatDamageBonus := map[int32]float64{
		25: 44,
		40: 80,
		50: 111,
		// TODO: AQ
		60: 138,
		// 60: 157
	}[warrior.Level]

	spellID := map[int32]int32{
		25: 1608,
		40: 11565,
		50: 11566,
		60: 11567,
	}[warrior.Level]

	// No known equation
	threat := map[int32]float64{
		25: 68,  //guess
		40: 103, //guess
		50: 120,
		60: 173,
	}[warrior.Level]

	warrior.HeroicStrike = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeMHAuto,
		Flags:       core.SpellFlagMeleeMetrics | core.SpellFlagNoOnCastComplete | SpellFlagBloodSurge,

		RageCost: core.RageCostOptions{
			Cost:   15 - float64(warrior.Talents.ImprovedHeroicStrike) - warrior.FocusedRageDiscount,
			Refund: 0.8,
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  threat,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower())
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

			if !result.Landed() {
				spell.IssueRefund(sim)
			}

			spell.DealDamage(sim, result)
			if warrior.curQueueAura != nil {
				warrior.curQueueAura.Deactivate(sim)
			}
		},
	})
	warrior.makeQueueSpellsAndAura(warrior.HeroicStrike)
}

func (warrior *Warrior) registerCleaveSpell() {
	flatDamageBonus := map[int32]float64{
		25: 5,
		40: 18,
		50: 32,
		60: 50,
	}[warrior.Level]

	spellID := map[int32]int32{
		25: 845,
		40: 11608,
		50: 11609,
		60: 20569,
	}[warrior.Level]

	threat := map[int32]float64{
		25: 20, //guess
		40: 60, //guess
		50: 80,
		60: 100,
	}[warrior.Level]

	flatDamageBonus *= []float64{1, 1.4, 1.8, 2.2}[warrior.Talents.ImprovedCleave]

	results := make([]*core.SpellResult, min(int32(2), warrior.Env.GetNumTargets()))

	warrior.Cleave = warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellID},
		SpellSchool: core.SpellSchoolPhysical,
		DefenseType: core.DefenseTypeMelee,
		ProcMask:    core.ProcMaskMeleeMHSpecial | core.ProcMaskMeleeMHAuto,
		Flags:       core.SpellFlagMeleeMetrics,

		RageCost: core.RageCostOptions{
			Cost: 20 - warrior.FocusedRageDiscount,
		},

		CritDamageBonus: warrior.impale(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		FlatThreatBonus:  threat,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			for idx := range results {
				baseDamage := flatDamageBonus + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower())
				results[idx] = spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
				target = sim.Environment.NextTargetUnit(target)
			}

			for _, result := range results {
				spell.DealDamage(sim, result)
			}

			if warrior.curQueueAura != nil {
				warrior.curQueueAura.Deactivate(sim)
			}
		},
	})
	warrior.makeQueueSpellsAndAura(warrior.Cleave)
}

func (warrior *Warrior) makeQueueSpellsAndAura(srcSpell *WarriorSpell) *WarriorSpell {
	queueAura := warrior.RegisterAura(core.Aura{
		Label:    "HS/Cleave Queue Aura-" + srcSpell.ActionID.String(),
		ActionID: srcSpell.ActionID.WithTag(1),
		Duration: core.NeverExpires,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			if warrior.curQueueAura != nil {
				warrior.curQueueAura.Deactivate(sim)
			}
			warrior.PseudoStats.DisableDWMissPenalty = true
			warrior.curQueueAura = aura
			warrior.curQueuedAutoSpell = srcSpell
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			warrior.PseudoStats.DisableDWMissPenalty = false
			warrior.curQueueAura = nil
			warrior.curQueuedAutoSpell = nil
		},
	})

	// Use a 50 ms ICD to simulate realistic delay between re-queueing HS
	icd := &core.Cooldown{
		Timer:    warrior.NewTimer(),
		Duration: core.SpellBatchWindow * 5,
	}

	queueSpell := warrior.RegisterSpell(AnyStance, core.SpellConfig{
		ActionID: srcSpell.ActionID.WithTag(1),
		Flags:    core.SpellFlagMeleeMetrics | core.SpellFlagAPL,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return icd.IsReady(sim) &&
				warrior.curQueueAura == nil &&
				warrior.CurrentRage() >= srcSpell.DefaultCast.Cost &&
				sim.CurrentTime >= warrior.Hardcast.Expires
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			queueAura.Activate(sim)
			icd.Use(sim)
		},
	})

	return queueSpell
}

func (warrior *Warrior) TryHSOrCleave(sim *core.Simulation, mhSwingSpell *core.Spell) *core.Spell {
	if !warrior.curQueueAura.IsActive() {
		return mhSwingSpell
	}

	if !warrior.curQueuedAutoSpell.CanCast(sim, warrior.CurrentTarget) {
		warrior.curQueueAura.Deactivate(sim)
		return mhSwingSpell
	}

	return warrior.curQueuedAutoSpell.Spell
}
