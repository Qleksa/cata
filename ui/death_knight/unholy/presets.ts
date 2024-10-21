import * as PresetUtils from '../../core/preset_utils';
import { Consumes, Flask, Food, Glyphs, Potions, Profession, PseudoStat, Stat, TinkerHands, UnitReference } from '../../core/proto/common';
import { DeathKnightMajorGlyph, DeathKnightMinorGlyph, DeathKnightPrimeGlyph, UnholyDeathKnight_Options } from '../../core/proto/death_knight';
import { SavedTalents } from '../../core/proto/ui';
import { Stats } from '../../core/proto_utils/stats';
import DefaultApl from '../../death_knight/unholy/apls/default.apl.json';
import SolaceApl from '../../death_knight/unholy/apls/solace.apl.json';
import P2BISGear from '../../death_knight/unholy/gear_sets/p2.bis.gear.json';
import P3BISGear from '../../death_knight/unholy/gear_sets/p3.bis.gear.json';
import PreBISGear from '../../death_knight/unholy/gear_sets/prebis.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.
export const PREBIS_GEAR_PRESET = PresetUtils.makePresetGear('Pre-bis', PreBISGear);
export const P2_BIS_GEAR_PRESET = PresetUtils.makePresetGear('P2 - BIS', P2BISGear);
export const P3_BIS_GEAR_PRESET = PresetUtils.makePresetGear('P3 - BIS', P3BISGear);

export const DEFAULT_ROTATION_PRESET = PresetUtils.makePresetAPLRotation('Default', DefaultApl);
export const SOLACE_SNAPSHOT_ROTATION_PRESET = PresetUtils.makePresetAPLRotation('Solace Snapshot', SolaceApl);

// Preset options for EP weights
export const P2_UNHOLY_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 4.49,
			[Stat.StatArmor]: 0.03,
			[Stat.StatAttackPower]: 1,
			[Stat.StatExpertiseRating]: 0.94,
			[Stat.StatHasteRating]: 2.4,
			[Stat.StatHitRating]: 2.6,
			[Stat.StatCritRating]: (1.43 + 0.69),
			[Stat.StatMasteryRating]: 1.65,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 6.13,
		},
	),
);

export const P3_UNHOLY_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 4.28,
			[Stat.StatArmor]: 0.03,
			[Stat.StatAttackPower]: 1,
			[Stat.StatExpertiseRating]: 1.03,
			[Stat.StatHasteRating]: 2.38,
			[Stat.StatHitRating]: 2.62,
			[Stat.StatCritRating]: 2.33,
			[Stat.StatMasteryRating]: 1.87,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 6.40,
			[PseudoStat.PseudoStatSpellHitPercent]: 156.60,
			[PseudoStat.PseudoStatPhysicalHitPercent]: 320.54,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wotlk.wowhead.com/talent-calc and copy the numbers in the url.

export const DefaultTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: '2032-1-13300321230231021231',
		glyphs: Glyphs.create({
			prime1: DeathKnightPrimeGlyph.GlyphOfDeathCoil,
			prime2: DeathKnightPrimeGlyph.GlyphOfScourgeStrike,
			prime3: DeathKnightPrimeGlyph.GlyphOfRaiseDead,
			major1: DeathKnightMajorGlyph.GlyphOfPestilence,
			major2: DeathKnightMajorGlyph.GlyphOfBloodBoil,
			major3: DeathKnightMajorGlyph.GlyphOfAntiMagicShell,
			minor1: DeathKnightMinorGlyph.GlyphOfDeathSEmbrace,
			minor2: DeathKnightMinorGlyph.GlyphOfPathOfFrost,
			minor3: DeathKnightMinorGlyph.GlyphOfHornOfWinter,
		}),
	}),
};

export const PREBIS_PRESET = PresetUtils.makePresetBuild('Pre-bis', {
	gear: PREBIS_GEAR_PRESET,
	epWeights: P2_UNHOLY_EP_PRESET,
})

export const P2_PRESET = PresetUtils.makePresetBuild('P2', {
	gear: P2_BIS_GEAR_PRESET,
	epWeights: P2_UNHOLY_EP_PRESET,
})

export const P3_PRESET = PresetUtils.makePresetBuild('P3', {
	gear: P3_BIS_GEAR_PRESET,
	epWeights: P3_UNHOLY_EP_PRESET,
})

export const DefaultOptions = UnholyDeathKnight_Options.create({
	classOptions: {
		startingRunicPower: 55,
		petUptime: 1,
	},
	unholyFrenzyTarget: UnitReference.create(),
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
};

export const DefaultConsumes = Consumes.create({
	flask: Flask.FlaskOfTitanicStrength,
	food: Food.FoodBeerBasedCrocolisk,
	defaultPotion: Potions.GolembloodPotion,
	prepopPotion: Potions.GolembloodPotion,
	tinkerHands: TinkerHands.TinkerHandsSynapseSprings,
});
