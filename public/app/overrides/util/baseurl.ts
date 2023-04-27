/**
 basename returns the "href" value of the <base> tag if available
 otherwise it assumes it's running under root, and location.origin is used
 it also expects it will only be run in the browser
  */
function basename() {
  const base = document.querySelector('base') as HTMLBaseElement;
  if (!base) {
    return window.location.origin;
  }

  return base.href;
}

export default basename;
