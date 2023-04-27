import basename from './baseurl';

function mockSelector(href: string) {
  const base = document.createElement('base');
  base.href = href;
  return base;
}

describe('baseurl', () => {
  describe('no baseURL meta tag set', () => {
    it('defaults to window.locaion.host', () => {
      const got = basename();
      expect(got).toBe('http://localhost');
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

        const got = basename();
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

        const got = basename();
        expect(got).toBe('http://localhost:9999/pyroscope');
      });
    });
  });
});
