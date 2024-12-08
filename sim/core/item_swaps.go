package core

import (
	"time"

	"github.com/wowsims/cata/sim/core/proto"
	"github.com/wowsims/cata/sim/core/stats"
)

type OnSwapItem func(*Simulation, proto.ItemSlot)

const numberOfGearPieces = int32(17)

type ItemSwap struct {
	character       *Character
	onSwapCallbacks [numberOfGearPieces][]OnSwapItem

	mhCritMultiplier     float64
	ohCritMultiplier     float64
	rangedCritMultiplier float64

	// Which slots to actually swap.
	slots []proto.ItemSlot

	// Holds items that are currently not equipped
	unEquippedItems []Item
	swapped         bool
}

/**
 * TODO All the extra parameters here and the code in multiple places for handling the Weapon struct is really messy,
 * we'll need to figure out something cleaner as this will be quite error-prone
**/
func (character *Character) enableItemSwap(itemSwap *proto.ItemSwap, mhCritMultiplier float64, ohCritMultiplier float64, rangedCritMultiplier float64) {
	var slots []proto.ItemSlot
	hasHandsSwap := itemSwap.HandsItem != nil && itemSwap.HandsItem.Id != 0
	hasMhSwap := itemSwap.MhItem != nil && itemSwap.MhItem.Id != 0
	hasOhSwap := itemSwap.OhItem != nil && itemSwap.OhItem.Id != 0
	hasRangedSwap := itemSwap.RangedItem != nil && itemSwap.RangedItem.Id != 0
	hasTrinket1Swap := itemSwap.Trinket_1Item != nil && itemSwap.Trinket_1Item.Id != 0
	hasTrinket2Swap := itemSwap.Trinket_2Item != nil && itemSwap.Trinket_2Item.Id != 0

	swapItems := []Item{
		proto.ItemSlot_ItemSlotHands:    toItem(itemSwap.HandsItem),
		proto.ItemSlot_ItemSlotMainHand: toItem(itemSwap.MhItem),
		proto.ItemSlot_ItemSlotOffHand:  toItem(itemSwap.OhItem),
		proto.ItemSlot_ItemSlotRanged:   toItem(itemSwap.RangedItem),
		proto.ItemSlot_ItemSlotTrinket1: toItem(itemSwap.Trinket_1Item),
		proto.ItemSlot_ItemSlotTrinket2: toItem(itemSwap.Trinket_2Item),
	}

	has2H := swapItems[proto.ItemSlot_ItemSlotMainHand].HandType == proto.HandType_HandTypeTwoHand
	hasMh := character.HasMHWeapon()
	hasOh := character.HasOHWeapon()

	if hasHandsSwap {
		slots = append(slots, proto.ItemSlot_ItemSlotHands)
	}
	// Handle MH and OH together, because present MH + empty OH --> swap MH and unequip OH
	if hasMhSwap || (hasOhSwap && hasMh) {
		slots = append(slots, proto.ItemSlot_ItemSlotMainHand)
	}
	if hasOhSwap || (has2H && hasOh) {
		slots = append(slots, proto.ItemSlot_ItemSlotOffHand)
	}
	if hasRangedSwap {
		slots = append(slots, proto.ItemSlot_ItemSlotRanged)
	}
	if hasTrinket1Swap {
		slots = append(slots, proto.ItemSlot_ItemSlotTrinket1)
	}
	if hasTrinket2Swap {
		slots = append(slots, proto.ItemSlot_ItemSlotTrinket2)
	}

	if len(slots) == 0 {
		return
	}

	character.ItemSwap = ItemSwap{
		mhCritMultiplier:     mhCritMultiplier,
		ohCritMultiplier:     ohCritMultiplier,
		rangedCritMultiplier: rangedCritMultiplier,
		slots:                slots,
		unEquippedItems:      swapItems,
		swapped:              false,
	}
}

func (swap *ItemSwap) initialize(character *Character) {
	swap.character = character
}

func (character *Character) RegisterOnItemSwap(slots []proto.ItemSlot, callback OnSwapItem) {
	if character == nil || !character.ItemSwap.IsEnabled() {
		return
	}

	var slot proto.ItemSlot
	for _, slot = range slots {
		character.ItemSwap.onSwapCallbacks[slot] = append(character.ItemSwap.onSwapCallbacks[slot], callback)
	}
}

// Helper for handling Effects that use PPMManager to toggle the aura on/off
func (swap *ItemSwap) RegisterOnSwapItemForEffectWithPPMManager(effectID int32, ppm float64, ppmm *PPMManager, aura *Aura) {
	character := swap.character
	character.RegisterOnItemSwap([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, func(sim *Simulation, slot proto.ItemSlot) {
		procMask := character.GetProcMaskForEnchant(effectID)
		*ppmm = character.AutoAttacks.NewPPMManager(ppm, procMask)

		if ppmm.Chance(procMask) == 0 {
			aura.Deactivate(sim)
		} else {
			aura.Activate(sim)
		}
	})
}

// Helper for handling procs that use PPMManager to toggle the aura on/off
func (swap *ItemSwap) RegisterOnSwapItemUpdateProcMaskWithPPMManager(procMask ProcMask, ppm float64, ppmm *PPMManager) {
	character := swap.character
	character.RegisterOnItemSwap([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}, func(sim *Simulation, slot proto.ItemSlot) {
		*ppmm = character.AutoAttacks.NewPPMManager(ppm, procMask)
	})
}

// Helper for handling Item Effects that use the itemID to toggle the aura on and off
func (swap *ItemSwap) RegisterOnSwapItemForItemProcEffect(itemID int32, aura *Aura, slots []proto.ItemSlot) {
	character := swap.character
	character.RegisterOnItemSwap(slots, func(sim *Simulation, slot proto.ItemSlot) {
		procMask := character.GetProcMaskForItem(itemID)
		if procMask == ProcMaskUnknown {
			aura.Deactivate(sim)
		} else {
			if !aura.IsActive() {
				aura.Activate(sim)
				if aura.Icd != nil {
					aura.Icd.Use(sim)
				}
			}
		}
	})
}

// Helper for handling Enchant Effects that use the effectID to toggle the aura on and off
func (swap *ItemSwap) RegisterOnSwapItemForEnchantProcEffect(effectID int32, aura *Aura, slots []proto.ItemSlot) {
	character := swap.character
	character.RegisterOnItemSwap(slots, func(sim *Simulation, slot proto.ItemSlot) {
		procMask := character.GetProcMaskForEnchant(effectID)

		if procMask == ProcMaskUnknown {
			aura.Deactivate(sim)
		} else {
			aura.Activate(sim)
		}
	})
}

// Helper for handling Item On Use effects to set a 30s cd on the related spell.
func (swap *ItemSwap) RegisterOnSwapItemForItemOnUseEffect(itemID int32, slots []proto.ItemSlot) {
	character := swap.character
	character.RegisterOnItemSwap(slots, func(sim *Simulation, slot proto.ItemSlot) {
		idEquipped := character.Equipment[slot].ID
		spell := swap.character.GetSpell(ActionID{ItemID: itemID})
		if spell != nil {
			aura := character.GetAuraByID(spell.ActionID)
			if aura.IsActive() {
				aura.Deactivate(sim)
			}
			if idEquipped != itemID {
				spell.Flags |= SpellFlagSwapped
				return
			}
			spell.Flags &= ^SpellFlagSwapped
			idSwapped := swap.unEquippedItems[slot].ID
			if idSwapped == idEquipped && spell.CD.IsReady(sim) || idSwapped != idEquipped {
				spell.CD.Set(sim.CurrentTime + time.Second*30)
			}
		}
	})
}

// Helper for handling Enchant On Use effects to set a 30s cd on the related spell.
func (swap *ItemSwap) RegisterOnSwapItemForEnchantOnUseEffect(spell *Spell, slots []proto.ItemSlot) {
	character := swap.character
	character.RegisterOnItemSwap(slots, func(sim *Simulation, slot proto.ItemSlot) {
		if spell != nil {
			idEquipped := character.Equipment[slot].ID
			idSwapped := swap.unEquippedItems[slot].ID
			if idSwapped == idEquipped && spell.CD.IsReady(sim) || idSwapped != idEquipped {
				spell.CD.Set(sim.CurrentTime + time.Second*30)
			}
		}
	})
}

func (swap *ItemSwap) IsEnabled() bool {
	return swap.character != nil && len(swap.slots) > 0
}

func (swap *ItemSwap) IsSwapped() bool {
	return swap.swapped
}

func (swap *ItemSwap) GetUnequippedItem(slot proto.ItemSlot) *Item {
	if slot < 0 {
		panic("Not able to swap Item " + slot.String() + " not supported")
	}
	return &swap.unEquippedItems[slot]
}

func (swap *ItemSwap) CalcStatChanges(slots []proto.ItemSlot) stats.Stats {
	newStats := stats.Stats{}
	for _, slot := range slots {
		oldItemStats := swap.getItemStats(swap.character.Equipment[slot])
		newItemStats := swap.getItemStats(*swap.GetUnequippedItem(slot))
		newStats = newStats.Add(newItemStats.Subtract(oldItemStats))
	}

	return newStats
}

func (swap *ItemSwap) SwapItems(sim *Simulation, slots []proto.ItemSlot) {
	if !swap.IsEnabled() {
		return
	}

	character := swap.character

	meleeWeaponSwapped := false
	newStats := stats.Stats{}
	has2H := swap.GetUnequippedItem(proto.ItemSlot_ItemSlotMainHand).HandType == proto.HandType_HandTypeTwoHand
	isPrepull := sim.CurrentTime < 0
	for _, slot := range slots {
		if !isPrepull && slot < proto.ItemSlot_ItemSlotMainHand || slot > proto.ItemSlot_ItemSlotOffHand {
			continue
		}

		//will swap both on the MainHand Slot for 2H.
		if slot == proto.ItemSlot_ItemSlotOffHand && has2H {
			continue
		}

		if ok, swapStats := swap.swapItem(slot, has2H); ok {
			newStats = newStats.Add(swapStats)
			meleeWeaponSwapped = slot == proto.ItemSlot_ItemSlotMainHand || slot == proto.ItemSlot_ItemSlotOffHand || meleeWeaponSwapped

			for _, onSwap := range swap.onSwapCallbacks[slot] {
				onSwap(sim, slot)
			}
		}
	}

	if sim.Log != nil {
		sim.Log("Item Swap Stats: %v", newStats.FlatString())
	}

	character.AddStatsDynamic(sim, newStats)

	if character.AutoAttacks.AutoSwingMelee && meleeWeaponSwapped && sim.CurrentTime > 0 {
		character.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime, false)
	}

	// If GCD is ready then use the GCD, otherwise we assume it's being used along side a spell.
	if character.GCD.IsReady(sim) {
		newGCD := sim.CurrentTime + 1500*time.Millisecond
		character.SetGCDTimer(sim, newGCD)
	}

	swap.swapped = !swap.swapped
}

func (swap *ItemSwap) swapItem(slot proto.ItemSlot, has2H bool) (bool, stats.Stats) {
	oldItem := swap.character.Equipment[slot]
	newItem := swap.GetUnequippedItem(slot)

	swap.character.Equipment[slot] = *newItem
	oldItemStats := swap.getItemStats(oldItem)
	newItemStats := swap.getItemStats(*newItem)
	newStats := newItemStats.Subtract(oldItemStats)

	//2H will swap out the offhand also.
	if has2H && slot == proto.ItemSlot_ItemSlotMainHand {
		_, ohStats := swap.swapItem(proto.ItemSlot_ItemSlotOffHand, has2H)
		newStats = newStats.Add(ohStats)
	}

	swap.unEquippedItems[slot] = oldItem
	swap.swapWeapon(slot)

	return true, newStats
}

func (swap *ItemSwap) getItemStats(item Item) stats.Stats {
	return ItemEquipmentStats(item)
}

func (swap *ItemSwap) swapWeapon(slot proto.ItemSlot) {
	character := swap.character

	switch slot {
	case proto.ItemSlot_ItemSlotMainHand:
		if character.AutoAttacks.AutoSwingMelee {
			character.AutoAttacks.SetMH(character.WeaponFromMainHand(swap.mhCritMultiplier))
		}
	case proto.ItemSlot_ItemSlotOffHand:
		if character.AutoAttacks.AutoSwingMelee {
			weapon := character.WeaponFromOffHand(swap.ohCritMultiplier)
			character.AutoAttacks.SetOH(weapon)

			character.AutoAttacks.IsDualWielding = weapon.SwingSpeed != 0
			character.PseudoStats.CanBlock = character.OffHand().WeaponType == proto.WeaponType_WeaponTypeShield
		}
	case proto.ItemSlot_ItemSlotRanged:
		if character.AutoAttacks.AutoSwingRanged {
			character.AutoAttacks.SetRanged(character.WeaponFromRanged(swap.rangedCritMultiplier))
		}
	}
}

func (swap *ItemSwap) reset(sim *Simulation) {
	if !swap.IsEnabled() {
		return
	}

	if swap.IsSwapped() {
		swap.SwapItems(sim, swap.slots)
	}

	for _, slot := range swap.slots {
		for _, onSwap := range swap.onSwapCallbacks[slot] {
			onSwap(sim, slot)
		}
	}
}

func (swap *ItemSwap) doneIteration(sim *Simulation) {
	if !swap.IsEnabled() || !swap.IsSwapped() {
		return
	}
}

func toItem(itemSpec *proto.ItemSpec) Item {
	if itemSpec == nil || itemSpec.Id == 0 {
		return Item{}
	}

	return NewItem(ItemSpec{
		ID:           itemSpec.Id,
		Gems:         itemSpec.Gems,
		Enchant:      itemSpec.Enchant,
		RandomSuffix: itemSpec.RandomSuffix,
		Reforging:    itemSpec.Reforging,
	})
}
