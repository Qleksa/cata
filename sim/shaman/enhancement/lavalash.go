package enhancement

import (
	"time"

	"github.com/wowsims/cata/sim/core"
	"github.com/wowsims/cata/sim/core/proto"
	"github.com/wowsims/cata/sim/shaman"
)

func (enh *EnhancementShaman) getSearingFlamesMultiplier() float64 {
	return enh.SearingFlamesMultiplier + core.TernaryFloat64(enh.HasT12Enh2pc, 0.05, 0)
}

func (enh *EnhancementShaman) registerLavaLashSpell() {
	damageMultiplier := 2.6
	if enh.SelfBuffs.ImbueOH == proto.ShamanImbue_FlametongueWeapon {
		damageMultiplier *= 1.4
	}

	enh.LavaLash = enh.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 78146},
		SpellSchool:    core.SpellSchoolFire,
		ProcMask:       core.ProcMaskMeleeOHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: shaman.SpellMaskLavaLash,
		ManaCost: core.ManaCostOptions{
			BaseCost: 0.04,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			IgnoreHaste: true,
			CD: core.Cooldown{
				Timer:    enh.NewTimer(),
				Duration: time.Second * 10,
			},
		},
		DamageMultiplier: damageMultiplier,
		CritMultiplier:   enh.DefaultSpellCritMultiplier(),
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := spell.Unit.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower())
			searingFlamesBonus := 1.0

			var searingFlames *core.Dot

			if enh.Talents.SearingFlames > 0 {
				searingFlames = enh.SearingFlames.Dot(target)
				searingFlamesBonus += enh.getSearingFlamesMultiplier() * float64(searingFlames.GetStacks())
			}

			baseDamage *= searingFlamesBonus
			result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

			if !result.Landed() {
				return
			}

			if enh.Talents.SearingFlames > 0 {
				if searingFlames.GetStacks() > 0 {
					numberSpread := 0
					maxTargets := 4
					for _, otherTarget := range sim.Encounter.TargetUnits {
						if otherTarget != target {
							enh.FlameShock.Cast(sim, otherTarget)
							numberSpread++
						}

						if numberSpread >= maxTargets {
							return
						}
					}

					searingFlames.SetStacks(sim, 0)
				}
			}
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return enh.HasOHWeapon()
		},
	})
}

func (enh *EnhancementShaman) IsLavaLashCastable(sim *core.Simulation) bool {
	return enh.LavaLash.IsReady(sim)
}
