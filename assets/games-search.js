import Tags from 'https://cdn.jsdelivr.net/npm/bootstrap5-tags@1.7.15/tags.min.js'

(() => {

    let searchField = null;
    let tagsField = null;
    let games = [];

    window.addEventListener('load', () => {
        prepareGames();

        searchField = document.getElementById('games-search');
        searchField.addEventListener('keyup', doSearch);

        Tags.init('#games-tags');
        const tagsEl = document.getElementById('games-tags');
        tagsField = Tags.getInstance(tagsEl);
        tagsField.setConfig('onSelectItem', doSearch);
        tagsField.setConfig('onClearItem', doSearch);

        document.querySelectorAll('span.game-tag').forEach(tagEl => {
            const tagValue = tagEl.innerText.trim();
            tagEl.addEventListener('click', () => {
                if (!tagsField.getSelectedValues().includes(tagValue)) {
                    tagsField.addItem(tagValue);
                    doSearch();
                }
            });
        });

        loadSearchParams();
    });

    function prepareGames() {
        document.querySelectorAll('div.game').forEach(el => {
            const name = el.querySelector('a.game-name').innerText;
            const tags = [];
            el.querySelectorAll('span.game-tag').forEach(tagEl => {
                tags.push(tagEl.innerText.trim());
            });
            games.push({el, name, tags, visible: true});
        });
    }

    function doSearch() {
        const searchText = searchField.value;
        const searchTags = tagsField.getSelectedValues();

        search(searchText, searchTags);

        if (window.sessionStorage) {
            sessionStorage._games_search_text = searchText;
            sessionStorage._games_search_tags = JSON.stringify(searchTags);
        }
    }

    function loadSearchParams() {
        if (!window.sessionStorage) {
            return;
        }
        if (window.sessionStorage._games_search_text) {
            searchField.value = window.sessionStorage._games_search_text;
        }
        if (window.sessionStorage._games_search_tags) {
            JSON.parse(window.sessionStorage._games_search_tags).forEach(tag => tagsField.addItem(tag));
        }
        doSearch();
    }

    function search(text, tags) {
        const searchText = escapeRegExpChars(text.trim());
        const re = new RegExp(searchText, 'ig');

        const containerEl = document.getElementById('games-container');

        games.forEach(game => {
            if (game.visible) {
                containerEl.removeChild(game.el);
            }
        });

        let gameFound = false;

        games.forEach(game => {
            const textMatched = !text || game.name.match(re);

            let tagsCount = 0;
            tags.forEach(tag => {
                if (game.tags.includes(tag)) {
                    tagsCount++;
                }
            });
            let tagsMatched = !tags || tags.length === 0 || tagsCount === tags.length;

            if (textMatched && tagsMatched) {
                containerEl.appendChild(game.el);
                game.visible = true;
                gameFound = true;
            } else {
                game.visible = false;
            }
        });

        notFound(!gameFound);
    }

    function notFound(value) {
        const el = document.getElementById('games-not-found');
        if (value) {
            el.classList.remove('d-none');
        } else {
            el.classList.add('d-none');
        }
    }

    function escapeRegExpChars(str) {
        if (!str) {
            return '';
        }
        const escaped = str.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&');
        return escaped.replace(/\s+/g, '\\s+');
    }

})();
