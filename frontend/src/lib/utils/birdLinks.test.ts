import { describe, expect, it } from 'vitest';

import { getBirdSiteLink, normalizeThumbnailClickDestination } from './birdLinks';

describe('normalizeThumbnailClickDestination', () => {
  it('defaults invalid destinations to details', () => {
    expect(normalizeThumbnailClickDestination('somewhere')).toBe('details');
    expect(normalizeThumbnailClickDestination(undefined)).toBe('details');
  });
});

describe('getBirdSiteLink', () => {
  const utm = 'utm_source=birdnet-go&utm_medium=thumbnail';

  it('builds bird-site links for bird species', () => {
    expect(
      getBirdSiteLink('ebird', 'Cardinalis cardinalis', 'Northern Cardinal', 'norcar', utm)
    ).toBe('https://ebird.org/species/norcar?utm_source=birdnet-go&utm_medium=thumbnail');

    expect(
      getBirdSiteLink('allaboutbirds', 'Cardinalis cardinalis', 'Northern Cardinal', 'norcar', utm)
    ).toBe(
      'https://allaboutbirds.org/guide/Northern_Cardinal?utm_source=birdnet-go&utm_medium=thumbnail'
    );
  });

  it('falls back to Wikipedia for non-bird taxonomy codes on bird-only destinations', () => {
    expect(
      getBirdSiteLink('ebird', 'Dryophytes cinereus', 'Green Treefrog', 't-11043207', utm)
    ).toBe('https://wikipedia.org/wiki/Dryophytes_cinereus');

    expect(
      getBirdSiteLink('allaboutbirds', 'Dryophytes cinereus', 'Green Treefrog', 't-11043207', utm)
    ).toBe('https://wikipedia.org/wiki/Dryophytes_cinereus');
  });

  it('falls back to Wikipedia for known non-bird sound classes', () => {
    expect(getBirdSiteLink('ebird', 'Dog', 'Dog', 'dogdog', utm)).toBe(
      'https://wikipedia.org/wiki/Dog'
    );
  });

  it('does not add UTM parameters to Wikipedia links', () => {
    expect(
      getBirdSiteLink('wikipedia', 'Dryophytes cinereus', 'Green Treefrog', 't-11043207', utm)
    ).toBe('https://wikipedia.org/wiki/Dryophytes_cinereus');
  });
});
