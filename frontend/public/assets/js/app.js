// public/js/app.js

document.addEventListener('alpine:init', () => {
    const components = {
      'dm-chat': '../dm',
      'room-chat': '../room',
      'track-manager': '../tracks',
      'product-manager': '../products',
      'product-docs': '../docs',
      'shared-ressources': '../shared',
      'tag-list': '../tags',
      'suggestion-panel': '../suggestions',
      'settings-panel': '../settings',
    };
  
    const loadComponent = async (id, url) => {
      const el = document.getElementById(id);
      if (el && el.innerHTML.trim() === '') {
        try {
          const res = await fetch(url);
          if (!res.ok) throw new Error(`Échec du chargement de ${url}`);
          const html = await res.text();
          el.innerHTML = html;
          evalScripts(el);
        } catch (e) {
          el.innerHTML = `<div class="text-red-500">Erreur de chargement du composant : ${id}</div>`;
          console.error(e);
        }
      }
    };
  
    const observer = new MutationObserver(() => {
      for (const [id, url] of Object.entries(components)) {
        const el = document.getElementById(id);
        if (el && el.offsetParent !== null) {
          loadComponent(id, url);
        }
      }
    });
  
    observer.observe(document.body, { childList: true, subtree: true });
  
    // Permet d'évaluer les balises <script> injectées dynamiquement
    function evalScripts(container) {
      container.querySelectorAll('script').forEach(script => {
        const s = document.createElement('script');
        if (script.src) {
          s.src = script.src;
        } else {
          s.textContent = script.textContent;
        }
        document.head.appendChild(s).parentNode.removeChild(s);
      });
    }
  });
  