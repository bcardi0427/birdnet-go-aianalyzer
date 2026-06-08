export type ThumbnailClickDestination =
  | 'details'
  | 'ebird'
  | 'wikipedia'
  | 'allaboutbirds'
  | 'none';

const VALID_THUMBNAIL_DESTINATIONS = new Set<ThumbnailClickDestination>([
  'details',
  'ebird',
  'wikipedia',
  'allaboutbirds',
  'none',
]);

const BIRD_ONLY_DESTINATIONS = new Set<ThumbnailClickDestination>(['ebird', 'allaboutbirds']);

// BirdNET includes some non-bird sound classes and taxa. Bird-only sites either
// 404 these pages or do not have useful guide pages, so send those to Wikipedia.
const NON_AVIAN_CLASSES = new Set([
  'siren',
  'dog',
  'power tools',
  'human vocal',
  'human non-vocal',
  'human whistle',
  'jet',
  'gun',
  'fireworks',
  'noise',
  'environmental',
  'engine',
]);

export function normalizeThumbnailClickDestination(
  destination: string | null | undefined
): ThumbnailClickDestination {
  if (destination && VALID_THUMBNAIL_DESTINATIONS.has(destination as ThumbnailClickDestination)) {
    return destination as ThumbnailClickDestination;
  }

  return 'details';
}

function normalizeUtmParameters(utmParameters: string | null | undefined): string {
  return (utmParameters ?? '').trim().replace(/^\?/, '').replace(/^&/, '');
}

function appendUtmParameters(url: string, utmParameters: string | null | undefined): string {
  const normalizedUtm = normalizeUtmParameters(utmParameters);
  if (!normalizedUtm) return url;

  return `${url}${url.includes('?') ? '&' : '?'}${normalizedUtm}`;
}

function getWikipediaLink(scientificName: string | null | undefined): string | null {
  if (!scientificName) return null;

  return `https://wikipedia.org/wiki/${encodeURIComponent(scientificName.replace(/ /g, '_'))}`;
}

function normalizeLabel(value: string | null | undefined): string {
  return (value ?? '').trim().toLowerCase();
}

function isNonAvianSpecies(
  scientificName: string | null | undefined,
  commonName: string | null | undefined,
  speciesCode: string | null | undefined
): boolean {
  const normalizedSpeciesCode = normalizeLabel(speciesCode);

  if (normalizedSpeciesCode.startsWith('t-')) {
    return true;
  }

  return (
    NON_AVIAN_CLASSES.has(normalizeLabel(scientificName)) ||
    NON_AVIAN_CLASSES.has(normalizeLabel(commonName))
  );
}

export function getBirdSiteLink(
  destination: string | null | undefined,
  scientificName: string | null | undefined,
  commonName: string | null | undefined,
  speciesCode: string | null | undefined,
  utmParameters: string | null | undefined
): string | null {
  const normalizedDestination = normalizeThumbnailClickDestination(destination);

  if (
    BIRD_ONLY_DESTINATIONS.has(normalizedDestination) &&
    isNonAvianSpecies(scientificName, commonName, speciesCode)
  ) {
    return getWikipediaLink(scientificName);
  }

  if (normalizedDestination === 'wikipedia') {
    return getWikipediaLink(scientificName);
  }

  if (normalizedDestination === 'allaboutbirds' && commonName) {
    const cleanedCommon = commonName.replace(/'/g, '');
    return appendUtmParameters(
      `https://allaboutbirds.org/guide/${encodeURIComponent(cleanedCommon.replace(/ /g, '_'))}`,
      utmParameters
    );
  }

  if (normalizedDestination === 'ebird' && speciesCode) {
    return appendUtmParameters(
      `https://ebird.org/species/${encodeURIComponent(speciesCode)}`,
      utmParameters
    );
  }

  return null;
}
