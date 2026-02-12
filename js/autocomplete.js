/**
 * Reusable autocomplete module
 *
 * Usage:
 * ```js
 * import { createAutocomplete } from './autocomplete.js';
 *
 * createAutocomplete({
 *   inputElement: document.querySelector('#search'),
 *   containerElement: document.querySelector('.search-form'),
 *   data: myDataArray,
 *   searchKey: 'search',
 *   displayKey: 'display',
 *   onSelect: (item) => console.log('Selected:', item),
 *   maxResults: 10,
 *   normalizeText: (text) => text.toLowerCase()
 * });
 * ```
 */

export function normalizeText(text) {
  // Match backend normalization: only remove specific Catalan accents.
  // Leave ç unchanged (unlike NFD normalization which converts ç to c).
  return text
    .replace(/[à]/g, "a")
    .replace(/[èé]/g, "e")
    .replace(/[íï]/g, "i")
    .replace(/[òó]/g, "o")
    .replace(/[úü]/g, "u");
}

export function createAutocomplete(options) {
  const {
    inputElement,
    containerElement,
    data,
    searchKey,
    displayKey,
    titleKey,
    onSelect,
    onClose,
    maxResults = 10,
    normalizeText: customNormalize = normalizeText,
    filterFunction,
    sortFunction,
    enableTextSelect = false,
  } = options;

  if (!inputElement || !containerElement || !data || !onSelect) {
    throw new Error("createAutocomplete requires: inputElement, containerElement, data, and onSelect");
  }

  const defaultFilterFunction = (items, _searchTerm, searchTermNormalized) =>
    items.filter((entry) => entry[searchKey].includes(searchTermNormalized));

  const defaultSortFunction = (items, searchTerm, searchTermNormalized) =>
    items.sort((a, b) => {
      // Check if starts with search term (exact accents).
      const aStartsExact = a[titleKey].startsWith(searchTerm);
      const bStartsExact = b[titleKey].startsWith(searchTerm);

      // Check if starts with search term (normalized, no accents).
      const aStarts = a[titleKey].startsWith(searchTermNormalized);
      const bStarts = b[titleKey].startsWith(searchTermNormalized);

      // 1. Prioritize exact accent matches that start with search term.
      if (aStartsExact && !bStartsExact) {
        return -1;
      }
      if (!aStartsExact && bStartsExact) {
        return 1;
      }

      // 2. Then prioritize normalized matches that start with search term.
      if (aStarts && !bStarts) {
        return -1;
      }
      if (!aStarts && bStarts) {
        return 1;
      }

      // 3. If both start or both don't start, maintain original order.
      return 0;
    });

  const autocompleteContainer = document.createElement("div");
  autocompleteContainer.className = "autocomplete-container";
  autocompleteContainer.tabIndex = -1;
  containerElement.appendChild(autocompleteContainer);

  // Track selected index for arrow key navigation.
  let selectedIndex = -1;

  // Track whether mouse has moved (to prevent hover selection on stationary mouse).
  let mouseHasMoved = true;

  // Track if we should select all text on focus (for mobile UX).
  let shouldSelectOnFocus = true;

  function closeAutocomplete() {
    autocompleteContainer.innerHTML = "";
    autocompleteContainer.classList.remove("active");
    selectedIndex = -1;
    if (onClose) {
      onClose();
    }
  }

  function updateSelectedItem(items, index) {
    for (const [i, item] of items.entries()) {
      if (i === index) {
        item.classList.add("selected");
        item.scrollIntoView({ behavior: "instant", block: "nearest" });
      } else {
        item.classList.remove("selected");
      }
    }
  }

  function displayAutocomplete(matches, onItemHover) {
    if (matches.length === 0) {
      closeAutocomplete();
      return;
    }

    autocompleteContainer.innerHTML = "";
    autocompleteContainer.classList.add("active");

    for (const [index, entry] of matches.entries()) {
      const item = document.createElement("div");
      item.className = "autocomplete-item";
      item.innerHTML = entry[displayKey];

      item.addEventListener("click", () => {
        closeAutocomplete();
        onSelect(entry);
      });

      // Sync mouse hover with keyboard selection.
      item.addEventListener("mouseenter", () => {
        if (onItemHover) {
          onItemHover(index);
        }
      });

      autocompleteContainer.appendChild(item);
    }
  }

  autocompleteContainer.addEventListener("mousemove", () => {
    mouseHasMoved = true;
  });

  // Select text when focusing the input (for easy replacement, especially on mobile).
  if (enableTextSelect) {
    inputElement.addEventListener("focus", () => {
      if (shouldSelectOnFocus && inputElement.value) {
        setTimeout(() => inputElement.select(), 0);
      }
    });

    inputElement.addEventListener("blur", () => {
      shouldSelectOnFocus = true;
    });
  }

  inputElement.addEventListener("input", (e) => {
    shouldSelectOnFocus = false;
    const searchTerm = e.target.value.trim().toLocaleLowerCase("ca");
    const searchTermNormalized = customNormalize(searchTerm);

    // Clear any visual selection and reset selected index when input changes.
    const previousItems = autocompleteContainer.querySelectorAll(".autocomplete-item");
    for (const item of previousItems) {
      item.classList.remove("selected");
    }
    selectedIndex = -1;

    // Reset mouse tracking when results change.
    mouseHasMoved = false;

    if (searchTerm.length === 0) {
      closeAutocomplete();
      return;
    }

    const filter = filterFunction || defaultFilterFunction;
    const sort = sortFunction || defaultSortFunction;

    let matches = filter(data, searchTerm, searchTermNormalized);
    matches = sort(matches, searchTerm, searchTermNormalized);
    matches = matches.slice(0, maxResults);

    displayAutocomplete(matches, (hoveredIndex) => {
      // Only update selection if mouse has actually moved.
      if (mouseHasMoved) {
        selectedIndex = hoveredIndex;
        const items = autocompleteContainer.querySelectorAll(".autocomplete-item");
        updateSelectedItem(items, selectedIndex);
      }
    });

    // Auto-select the first item by default.
    if (matches.length > 0) {
      selectedIndex = 0;
      const items = autocompleteContainer.querySelectorAll(".autocomplete-item");
      updateSelectedItem(items, selectedIndex);
    }
  });

  inputElement.addEventListener("keydown", (e) => {
    mouseHasMoved = false;
    const items = autocompleteContainer.querySelectorAll(".autocomplete-item");

    if (items.length === 0) {
      return;
    }

    if (e.key === "ArrowDown") {
      e.preventDefault();
      selectedIndex = (selectedIndex + 1) % items.length;
      updateSelectedItem(items, selectedIndex);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      selectedIndex = selectedIndex <= 0 ? items.length - 1 : selectedIndex - 1;
      updateSelectedItem(items, selectedIndex);
    } else if (e.key === "Enter") {
      e.preventDefault();
      // Click the selected item (first item is selected by default).
      if (selectedIndex >= 0 && items.length > 0) {
        items[selectedIndex].click();
      }
    } else if (e.key === "Escape") {
      closeAutocomplete();
    }
  });

  document.addEventListener("click", (e) => {
    if (!containerElement.contains(e.target)) {
      closeAutocomplete();
    }
  });

  // Return API for programmatic control.
  return {
    close() {
      closeAutocomplete();
    },
    getContainer() {
      return autocompleteContainer;
    },
  };
}
