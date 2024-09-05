package druid

import (
	"fmt"
	"time"

	"github.com/wowsims/sod/sim/core"
)

const MoonfireRanks = 10

var MoonfireSpellId = [MoonfireRanks + 1]int32{0, 8921, 8924, 8925, 8926, 8927, 8928, 8929, 9833, 9834, 9835}
var MoonfiresSpellCoeff = [MoonfireRanks + 1]float64{0, .06, .094, .128, .15, .15, .15, .15, .15, .15, .15}
var MoonfiresSellDotCoeff = [MoonfireRanks + 1]float64{0, .052, .081, .111, .13, .13, .13, .13, .13, .13, .13}
var MoonfireBaseDamage = [MoonfireRanks + 1][]float64{{0}, {9, 12}, {17, 21}, {30, 37}, {44, 53}, {70, 82}, {91, 108}, {117, 137}, {143, 168}, {172, 200}, {195, 228}}
var MoonfireBaseDotDamage = [MoonfireRanks + 1]float64{0, 12, 32, 52, 80, 124, 164, 212, 264, 320, 384}
var MoonfireDotTicks = [MoonfireRanks + 1]int32{0, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4}
var MoonfireManaCost = [MoonfireRanks + 1]float64{0, 25, 50, 75, 105, 150, 190, 235, 280, 325, 375}
var MoonfireLevel = [MoonfireRanks + 1]int{0, 4, 10, 16, 22, 28, 34, 40, 46, 52, 58}

func (druid *Druid) registerMoonfireSpell() {
	druid.Moonfire = make([]*DruidSpell, 0)

	for rank := 1; rank <= MoonfireRanks; rank++ {
		config := druid.getMoonfireBaseConfig(rank)

		if config.RequiredLevel <= int(druid.Level) {
			druid.Moonfire = append(druid.Moonfire, druid.RegisterSpell(Humanoid|Moonkin, config))
		}
	}
}

func (druid *Druid) getMoonfireBaseConfig(rank int) core.SpellConfig {
	ticks := MoonfireDotTicks[rank]
	tickLength := time.Second * 3
	talentBaseMultiplier := 1 + druid.MoonfuryDamageMultiplier() + druid.ImprovedMoonfireDamageMultiplier()

	spellId := MoonfireSpellId[rank]
	spellCoeff := MoonfiresSpellCoeff[rank]
	spellDotCoeff := MoonfiresSellDotCoeff[rank]
	baseDamageLow := MoonfireBaseDamage[rank][0] * talentBaseMultiplier
	baseDamageHigh := MoonfireBaseDamage[rank][1] * talentBaseMultiplier
	baseDotDamage := (MoonfireBaseDotDamage[rank] / float64(ticks)) * talentBaseMultiplier
	manaCost := MoonfireManaCost[rank]
	level := MoonfireLevel[rank]

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_DruidMoonfire,
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagOmen | core.SpellFlagAPL | core.SpellFlagResetAttackSwing,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost:   manaCost,
			Multiplier: druid.MoonglowManaCostMultiplier(),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 0,
			},
		},
		Dot: core.DotConfig{
			Aura: core.Aura{
				Label:    fmt.Sprintf("Moonfire (Rank %d)", rank),
				ActionID: core.ActionID{SpellID: spellId},
			},
			NumberOfTicks:    ticks,
			TickLength:       tickLength,
			BonusCoefficient: spellDotCoeff,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				dot.Snapshot(target, baseDotDamage, isRollover)
				dot.SnapshotAttackerMultiplier *= druid.MoonfireDotMultiplier
				if !druid.form.Matches(Moonkin) {
					dot.SnapshotCritChance = 0
				}
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeSnapshotCrit)
			},
		},

		BonusCoefficient: spellCoeff,
		BonusCritRating:  druid.ImprovedMoonfireCritBonus(),
		CritDamageBonus:  druid.vengeanceBonusCritDamage(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

			if result.Landed() {
				dot := spell.Dot(target)
				dot.Apply(sim)
			}
		},
	}
}
