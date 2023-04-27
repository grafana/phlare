import baseurl from './baseurl';
import { detectBaseurl } from './baseurl';

function mockSelector(href: string) {
  const base = document.createElement('base');
  base.href = href;
  return base;
}

describe('baseurl', () => {
  describe('no baseurl has been detected', () => {
    it('returns undefined', () => {
      const got = baseurl();
      expect(got).toBe(undefined);
    });
  });

  describe('base tag is set', () => {
    describe('it contains /ui', () => {
      it('removes /ui path', () => {
        jest
          .spyOn(document, 'querySelector')
          .mockImplementationOnce(() => mockSelector('/pyroscope/ui'));

        const got = baseurl();
        expect(got).toBe('/pyroscope');
      });
    });

    describe('it does not contain /ui', () => {
      // Until we move the /ui page to the root, this is absolutely required
      it('throws an error', () => {
        jest
          .spyOn(document, 'querySelector')
          .mockImplementationOnce(() => mockSelector('/pyroscope/'));

        expect(() => baseurl()).toThrowError();
      });
    });
  });
});

describe('detectBaseurl', () => {
  describe('no baseURL meta tag set', () => {
    it('returns undefined', () => {
      const got = detectBaseurl();
      expect(got).toBe(undefined);
    });
  });

  describe('base tag is set', () => {
    describe('a relative path is passed', () => {
      // This test ends up testing some functionality for 'base' tag
      // which is left for documentating purposes
      it('prepends with the host', () => {
        jest
          .spyOn(document, 'querySelector')
          .mockImplementationOnce(() => mockSelector('/pyroscope'));

        const got = detectBaseurl();
        expect(got).toBe('http://localhost/pyroscope');
      });
    });

    describe('a full url is passed', () => {
      it('uses as is', () => {
        jest
          .spyOn(document, 'querySelector')
          .mockImplementationOnce(() =>
            mockSelector('http://localhost:9999/pyroscope')
          );

        const got = detectBaseurl();
        expect(got).toBe('http://localhost:9999/pyroscope');
      });
    });
  });
});
