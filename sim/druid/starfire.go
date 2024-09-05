package druid

import (
	"time"

	item_sets "github.com/wowsims/sod/sim/common/sod/items_sets"
	"github.com/wowsims/sod/sim/core"
)

const StarfireRanks = 7

var StarfireSpellId = [StarfireRanks + 1]int32{0, 2912, 8949, 8950, 8951, 9875, 9876, 25298}
var StarfireBaseDamage = [StarfireRanks + 1][]float64{{0}, {95, 115}, {146, 177}, {212, 253}, {293, 348}, {378, 445}, {451, 531}, {496, 584}}
var StarfireManaCost = [StarfireRanks + 1]float64{0, 95, 135, 180, 230, 275, 315, 340}
var StarfireLevel = [StarfireRanks + 1]int{0, 20, 26, 34, 42, 50, 58, 60}

func (druid *Druid) registerStarfireSpell() {
	druid.Starfire = make([]*DruidSpell, StarfireRanks+1)

	// TODO: AQ <=
	for rank := 1; rank < StarfireRanks; rank++ {
		config := druid.newStarfireSpellConfig(rank)

		if config.RequiredLevel <= int(druid.Level) {
			druid.Starfire[rank] = druid.RegisterSpell(Humanoid|Moonkin, config)
		}
	}
}

func (druid *Druid) newStarfireSpellConfig(rank int) core.SpellConfig {
	talentBaseMultiplier := 1 + druid.MoonfuryDamageMultiplier()

	spellId := StarfireSpellId[rank]
	baseDamageLow := StarfireBaseDamage[rank][0] * talentBaseMultiplier
	baseDamageHigh := StarfireBaseDamage[rank][1] * talentBaseMultiplier
	manaCost := StarfireManaCost[rank]
	level := StarfireLevel[rank]

	castTime := 3500

	return core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellId},
		SpellCode:   SpellCode_DruidStarfire,
		SpellSchool: core.SpellSchoolArcane,
		DefenseType: core.DefenseTypeMagic,
		ProcMask:    core.ProcMaskSpellDamage,
		Flags:       SpellFlagOmen | core.SpellFlagAPL | core.SpellFlagResetAttackSwing,

		RequiredLevel: level,
		Rank:          rank,

		ManaCost: core.ManaCostOptions{
			FlatCost: manaCost * (1 - 0.03*float64(druid.Talents.Moonglow)),
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: time.Millisecond*time.Duration(castTime) - time.Millisecond*100*time.Duration(druid.Talents.ImprovedStarfire),
			},
		},

		BonusCritRating: core.TernaryFloat64(druid.HasSetBonus(item_sets.ItemSetInsulatedSorcerorLeather, 3), 2, 0) * core.CritRatingPerCritChance,

		CritDamageBonus: druid.vengeanceBonusCritDamage(),

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := sim.Roll(baseDamageLow, baseDamageHigh)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		},
	}
}
