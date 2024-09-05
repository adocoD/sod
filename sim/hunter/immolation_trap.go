package hunter

import (
	"strconv"
	"time"

	"github.com/wowsims/sod/sim/core"
)

func (hunter *Hunter) getImmolationTrapConfig(rank int, timer *core.Timer) core.SpellConfig {
	spellId := [6]int32{0, 409521, 409524, 409526, 409528, 409530}[rank]
	dotDamage := [6]float64{0, 105, 215, 340, 510, 690}[rank]
	manaCost := [6]float64{0, 50, 90, 135, 190, 245}[rank]
	level := [6]int{0, 16, 26, 36, 46, 56}[rank]

	return core.SpellConfig{
		ActionID:      core.ActionID{SpellID: spellId},
		SpellSchool:   core.SpellSchoolFire,
		DefenseType:   core.DefenseTypeMagic,
		ProcMask:      core.ProcMaskSpellDamage,
		Flags:         core.SpellFlagAPL | core.SpellFlagPassiveSpell | SpellFlagTrap,
		Rank:          rank,
		RequiredLevel: level,
		MissileSpeed:  24,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost * hunter.resourcefulnessManacostModifier(),
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    timer,
				Duration: time.Second * time.Duration(15*hunter.resourcefulnessCooldownModifier()),
			},
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true, // Hunter GCD is locked at 1.5s
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.DistanceFromTarget <= hunter.trapRange()
		},

		BonusHitRating: hunter.trapMastery(),

		DamageMultiplier: (1 + 0.15*float64(hunter.Talents.CleverTraps)) * hunter.tntDamageMultiplier(),
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				Label: "ImmolationTrap" + hunter.Label + strconv.Itoa(rank),
				Tag:   "ImmolationTrap",
			},
			NumberOfTicks: 5,
			TickLength:    time.Millisecond * 1500,

			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				tickDamage := (dotDamage + hunter.tntDamageFlatBonus()) / float64(dot.NumberOfTicks)
				dot.Snapshot(target, tickDamage, isRollover)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTick)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealOutcome(sim, result)
				if result.Landed() {
					spell.Dot(target).Apply(sim)
				}
			})
		},
	}
}

func (hunter *Hunter) registerImmolationTrapSpell(timer *core.Timer) {
	maxRank := 5
	for i := 1; i <= maxRank; i++ {
		config := hunter.getImmolationTrapConfig(i, timer)

		if config.RequiredLevel <= int(hunter.Level) {
			hunter.ImmolationTrap = hunter.GetOrRegisterSpell(config)
		}
	}
}
