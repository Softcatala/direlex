import { createAutocomplete } from "./autocomplete.js";
import terms from "./data/terms.json" with { type: "json" };

const searchInput = document.querySelector('input[type="search"]');
const isMobile = /Android|iPad|iPhone/i.test(navigator.userAgent);

const autocompleteInstance = createAutocomplete({
  containerElement: document.querySelector(".search-form"),
  inputElement: searchInput,
  data: terms,
  displayKey: "d",
  searchKey: "s",
  titleKey: "t",
  onSelect: (entry) => {
    searchInput.value = entry.t.replaceAll("_", " ");
    // Match Go template's URL encoding behavior:
    // - Encode: ( ) ' (encodeURIComponent leaves these unescaped)
    // - Decode: [ ] : (encodeURIComponent encodes these but Go doesn't)
    // CMS content links may use PHP urlencode output (encodes [ ] :), which still resolves.
    // Current data slugs include only these reserved characters: [ ] : ( ) ' |.
    // If slugs ever include other reserved URL characters, keep Go/JS encoders in sync.
    // Note: Hex digits case differs (JS: uppercase, Go template: lowercase)
    // but this is functionally equivalent per RFC 3986.
    const slug = encodeURIComponent(entry.t)
      .replaceAll("(", "%28")
      .replaceAll(")", "%29")
      .replaceAll("'", "%27")
      .replaceAll("%5B", "[")
      .replaceAll("%5D", "]")
      .replaceAll("%3A", ":");
    window.location.href = `/lema/${slug}`;
  },
  enableTextSelect: isMobile,
});

window.addEventListener("pageshow", () => {
  // Ensure browser does not try to remember last form value.
  const lema = location.pathname.startsWith("/lema/") ? decodeURIComponent(location.pathname.slice(6)) : "";
  searchInput.value = lema.replaceAll("_", " ");
  autocompleteInstance.close();

  if (!isMobile) {
    // On desktop, select the searched value so it can be replaced by typing.
    searchInput.select();
  }
});
