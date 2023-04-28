/**
 basename returns the "href" value of the <base> tag if available
 otherwise it assumes it's running under root, and location.origin is used
 it also expects it will only be run in the browser
 */
export function baseurl() {
  const href = detectBaseurl();
  if (!href) {
    return undefined;
  }

  const url = new URL(href, window.location.href);
  return url.pathname;
}

export function baseurlForAPI() {
  // When serving production, api path is one level above /ui
  return baseurl()?.replace('/ui', '');
}

function detectBaseurl(): string | undefined {
  const base = document.querySelector('base') as HTMLBaseElement;
  if (!base) {
    return undefined;
  }

  return base.href;
}

export default baseurlForAPI;
