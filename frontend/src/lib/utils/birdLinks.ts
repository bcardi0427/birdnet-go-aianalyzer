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

export function getBirdSiteLink(
  destination: string | null | undefined,
  scientificName: string | null | undefined,
  commonName: string | null | undefined,
  speciesCode: string | null | undefined,
  utmParameters: string | null | undefined
): string | null {
  const normalizedDestination = normalizeThumbnailClickDestination(destination);

  if (normalizedDestination === 'wikipedia' && scientificName) {
    return `https://wikipedia.org/wiki/${encodeURIComponent(scientificName.replace(/ /g, '_'))}`;
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
