import { createAutocomplete } from "./autocomplete.js";
import terms from "./data/terms-glossary.json" with { type: "json" };

const searchInput = document.querySelector('input[type="search"]');
const isMobile = /Android|iPad|iPhone/i.test(navigator.userAgent);

const autocompleteInstance = createAutocomplete({
  containerElement: document.querySelector(".search-form"),
  inputElement: searchInput,
  data: terms,
  displayKey: "d",
  searchKey: "s",
  titleKey: "id",
  onSelect: (entry) => {
    searchInput.value = "";
    window.location.href = `#${entry.id}`;
  },
  enableTextSelect: isMobile,
});

window.addEventListener("pageshow", () => {
  // Ensure browser does not try to remember last form value.
  searchInput.value = "";
  autocompleteInstance.close();
});

const backToTop = document.getElementById("back-to-top");
const scrollThreshold = 300;

setInterval(() => {
  if (window.scrollY > scrollThreshold) {
    backToTop.classList.add("visible");
  } else {
    backToTop.classList.remove("visible");
  }
}, 250);

backToTop.addEventListener("click", (e) => {
  e.preventDefault();
  document.querySelector("section.content").scrollIntoView({ behavior: "instant" });
  if (!isMobile) {
    searchInput.focus();
  }
});
