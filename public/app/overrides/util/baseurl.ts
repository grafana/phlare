/**
 * basename returns the baseurl to be used when requesting the backend
 * it first detects the href from <base> tag, then removes the /ui subpath
 */
export function baseurl() {
  const href = detectBaseurl();
  if (!href) {
    return undefined;
  }

  const url = new URL(href, window.location.href);
  if (!url.pathname.includes('/ui')) {
    throw new Error('/ui path under <base> tag is expected');
  }
  return url.pathname.replace('/ui', '');
}

/**
 basename returns the "href" value of the <base> tag if available
 otherwise it assumes it's running under root, and location.origin is used
 it also expects it will only be run in the browser
  */
export function detectBaseurl(): string | undefined {
  const base = document.querySelector('base') as HTMLBaseElement;
  if (!base) {
    return undefined;
  }

  return base.href;
}

export default baseurl;
