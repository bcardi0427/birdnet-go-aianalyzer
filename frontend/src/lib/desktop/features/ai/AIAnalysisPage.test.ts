import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import AIAnalysisPage from './AIAnalysisPage.svelte';

const getReportMock = vi.fn();

vi.mock('$lib/utils/settingsApi', () => ({
  settingsAPI: {
    ai: {
      getReport: () => getReportMock(),
    },
  },
}));

vi.mock('$lib/i18n', () => ({
  t: (key: string) => key,
}));

vi.mock('$lib/stores/navigation.svelte', () => ({
  navigation: {
    navigate: vi.fn(),
  },
}));

describe('AIAnalysisPage', () => {
  beforeEach(() => {
    getReportMock.mockReset();
  });

  it('renders report content and strips non-local image sources', async () => {
    getReportMock.mockResolvedValue({
      report:
        '# Test\n\nSafe text\n\n<img src="https://example.com/bad.jpg" alt="bad">\n<img src="/api/v2/media/species-image?name=Corvus%20brachyrhynchos" alt="ok">',
      generatedAt: new Date().toISOString(),
      cached: true,
    });

    const { container } = render(AIAnalysisPage);

    await waitFor(() => {
      expect(screen.getByText('Safe text')).toBeInTheDocument();
    });

    const images = Array.from(container.querySelectorAll('img'));
    expect(images.length).toBe(1);
    expect(images[0].getAttribute('src')).toContain('/api/v2/media/species-image');
  });

  it('shows friendly error for missing API key', async () => {
    getReportMock.mockRejectedValue(new Error('Gemini API key is not configured'));

    render(AIAnalysisPage);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('AI API key is missing');
    });
  });
});
