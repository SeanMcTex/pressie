// Pressie SPA — vanilla JS, no build step.
// IA: contacts list → contact detail (ideas/given/received tabs)
// Progressive disclosure: add forms behind buttons, not always visible.

const API = '';
let currentRoute = '';

// --- Router ---
function router() {
  const hash = window.location.hash.slice(1) || '/';
  currentRoute = hash;
  const parts = hash.split('/').filter(Boolean);

  if (parts.length === 0) {
    renderContactsPage();
  } else if (parts[0] === 'contacts' && parts[1]) {
    renderContactPage(decodeURIComponent(parts[1]));
  } else if (parts[0] === 'ideas') {
    renderGeneralIdeasPage();
  } else {
    renderContactsPage();
  }
}

window.addEventListener('hashchange', router);

// --- API helpers ---
async function api(path, opts = {}) {
  const res = await fetch(API + path, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return res.json();
}

// --- Render helpers ---
function el(tag, props = {}, ...children) {
  const e = document.createElement(tag);
  for (const [k, v] of Object.entries(props)) {
    if (k === 'class') e.className = v;
    else if (k === 'onclick') e.onclick = v;
    else if (k === 'href') e.setAttribute('href', v);
    else if (k === 'type') e.type = v;
    else if (k === 'placeholder') e.placeholder = v;
    else if (k === 'value') e.value = v;
    else if (k === 'style') e.style.cssText = v;
    else if (k === 'id') e.id = v;
    else if (k === 'rows') e.rows = v;
    else if (k === 'target') e.setAttribute('target', v);
    else if (v !== undefined) e.setAttribute(k, v);
  }
  for (const child of children) {
    if (child == null) continue;
    e.appendChild(typeof child === 'string' ? document.createTextNode(child) : child);
  }
  return e;
}

function tagPill(text) {
  return el('span', { class: 'tag' }, text);
}

function statusBadge(status) {
  return el('span', { class: `badge badge-${status}` }, status);
}

function navigate(path) {
  window.location.hash = `/${path}`;
}

// --- Header (shared) ---
function header(title) {
  const h = el('div', { class: 'header' });
  if (currentRoute !== '/' && currentRoute !== '') {
    h.appendChild(el('a', { href: '#/', class: 'back-link' }, '← Back'));
  }
  h.appendChild(el('h1', {}, title));
  const nav = el('div', { class: 'nav' });
  nav.appendChild(el('a', { href: '#/', class: currentRoute === '/' ? 'active' : '' }, 'Contacts'));
  nav.appendChild(el('a', { href: '#/ideas', class: currentRoute.startsWith('/ideas') ? 'active' : '' }, 'General Ideas'));
  h.appendChild(nav);
  return h;
}

// --- Contacts page (home) ---

async function renderContactsPage() {
  const app = document.getElementById('app');
  app.innerHTML = '';

  // Header with "quick add" button
  const h = el('div', { class: 'header' },
    el('h1', {}, 'Contacts'),
    el('div', { class: 'nav' },
      el('a', { href: '#/', class: 'active' }, 'Contacts'),
      el('a', { href: '#/ideas' }, 'General Ideas'),
    ),
  );
  app.appendChild(h);

  // Quick-add toggle (hidden by default)
  const quickAddBtn = el('button', {
    class: 'secondary',
    onclick: () => {
      const form = document.getElementById('quick-add-form');
      form.style.display = form.style.display === 'none' ? 'block' : 'none';
    },
  }, '+ Quick add idea');
  app.appendChild(quickAddBtn);

  // Quick add form (hidden until toggled)
  const addForm = el('div', { class: 'add-form', id: 'quick-add-form', style: 'display:none;' },
    el('h3', {}, 'Quick add idea'),
    el('div', { class: 'form-group' },
      el('label', {}, 'For (name)'),
      el('input', { type: 'text', id: 'qa-for', placeholder: 'Kris' }),
    ),
    el('div', { class: 'form-group' },
      el('label', {}, 'Item'),
      el('input', { type: 'text', id: 'qa-item', placeholder: 'Letterpress print' }),
    ),
    el('div', { class: 'form-group' },
      el('label', {}, 'Tags (comma-separated)'),
      el('input', { type: 'text', id: 'qa-tags', placeholder: 'art, irish' }),
    ),
    el('button', { onclick: quickAddIdea }, 'Add idea'),
  );
  app.appendChild(addForm);

  // Contacts list — front and center
  const listDiv = el('div', { class: 'mt-2' });
  app.appendChild(listDiv);

  try {
    const contacts = await api('/api/contacts');
    if (contacts.length === 0) {
      listDiv.appendChild(el('div', { class: 'empty-state' },
        el('p', {}, 'No contacts yet.'),
        el('p', {}, 'Add an idea with the "+ Quick add idea" button above.'),
      ));
      return;
    }
    for (const c of contacts) {
      const item = el('div', { class: 'contact-item', onclick: () => navigate(`contacts/${c.name}`) },
        el('div', {},
          el('div', { class: 'name' }, c.name),
          c.tags && c.tags.length > 0 ? el('div', { class: 'tags' }, ...c.tags.map(tagPill)) : null,
        ),
        el('span', { class: 'text-muted' }, c.visibility),
      );
      listDiv.appendChild(item);
    }
  } catch (err) {
    listDiv.appendChild(el('div', { class: 'empty-state' }, el('p', {}, `Error: ${err.message}`)));
  }
}

// --- Contact detail page ---

async function renderContactPage(name) {
  const app = document.getElementById('app');
  app.innerHTML = '';

  app.appendChild(header(name));

  try {
    const contact = await api(`/api/contacts/${encodeURIComponent(name)}`);

    // Preferences (read-only summary, edit behind a button)
    if (contact.preferences) {
      app.appendChild(el('div', { class: 'preferences-box' },
        el('div', { class: 'label' }, 'Preferences'),
        el('div', {}, contact.preferences),
      ));
    }

    // Tabs: Ideas / Given / Received
    const tabContainer = el('div', {});
    const tabs = el('div', { class: 'tabs' },
      el('div', { class: 'tab active', onclick: (e) => showTab(tabContainer, 'ideas', contact, e.target) }, 'Ideas'),
      el('div', { class: 'tab', onclick: (e) => showTab(tabContainer, 'given', contact, e.target) }, 'Given'),
      el('div', { class: 'tab', onclick: (e) => showTab(tabContainer, 'received', contact, e.target) }, 'Received'),
    );
    app.appendChild(tabs);
    app.appendChild(tabContainer);
    showTab(tabContainer, 'ideas', contact);

    // Action buttons (progressive disclosure)
    const actionsRow = el('div', { class: 'actions-row mt-2' },
      el('button', {
        class: 'secondary',
        onclick: () => toggleForm('idea-form'),
      }, '+ Add idea'),
      el('button', {
        class: 'secondary',
        onclick: () => toggleForm('gift-form'),
      }, '+ Log gift'),
      el('button', {
        class: 'secondary',
        onclick: () => toggleForm('prefs-form'),
      }, contact.preferences ? 'Edit preferences' : 'Set preferences'),
    );
    app.appendChild(actionsRow);

    // Add idea form (hidden)
    app.appendChild(buildIdeaForm(name));

    // Log gift form (hidden)
    app.appendChild(buildGiftForm(name));

    // Preferences form (hidden)
    app.appendChild(buildPrefsForm(name, contact.preferences));

  } catch (err) {
    app.appendChild(el('div', { class: 'empty-state' }, el('p', {}, `Error: ${err.message}`)));
  }
}

function toggleForm(id) {
  const form = document.getElementById(id);
  if (!form) return;
  // Hide all other forms
  ['idea-form', 'gift-form', 'prefs-form'].forEach(fid => {
    if (fid !== id) {
      const f = document.getElementById(fid);
      if (f) f.style.display = 'none';
    }
  });
  form.style.display = form.style.display === 'none' ? 'block' : 'none';
}

function buildIdeaForm(name) {
  return el('div', { class: 'add-form', id: 'idea-form', style: 'display:none;' },
    el('h3', {}, 'Add idea'),
    el('div', { class: 'form-group' },
      el('label', {}, 'Item'),
      el('input', { type: 'text', id: 'add-item', placeholder: 'Gift idea description' }),
    ),
    el('div', { class: 'form-row' },
      el('div', { class: 'form-group' },
        el('label', {}, 'Tags'),
        el('input', { type: 'text', id: 'add-tags', placeholder: 'art, irish' }),
      ),
      el('div', { class: 'form-group' },
        el('label', {}, 'URL'),
        el('input', { type: 'text', id: 'add-url', placeholder: 'https://...' }),
      ),
    ),
    el('div', { class: 'form-group' },
      el('label', {}, 'Notes'),
      el('input', { type: 'text', id: 'add-notes', placeholder: 'Notes' }),
    ),
    el('button', { onclick: () => addIdea(name) }, 'Add idea'),
  );
}

function buildGiftForm(name) {
  return el('div', { class: 'add-form', id: 'gift-form', style: 'display:none;' },
    el('h3', {}, 'Log a gift'),
    el('div', { class: 'form-group' },
      el('label', {}, 'Item'),
      el('input', { type: 'text', id: 'gift-item', placeholder: 'Gift description' }),
    ),
    el('div', { class: 'form-row' },
      el('div', { class: 'form-group' },
        el('label', {}, 'Occasion'),
        el('input', { type: 'text', id: 'gift-occasion', placeholder: 'christmas' }),
      ),
      el('div', { class: 'form-group' },
        el('label', {}, 'Date'),
        el('input', { type: 'text', id: 'gift-date', placeholder: '2025-12-25' }),
      ),
    ),
    el('div', { class: 'form-row' },
      el('div', { class: 'form-group' },
        el('label', {}, 'Price'),
        el('input', { type: 'number', id: 'gift-price', placeholder: '50' }),
      ),
      el('div', { class: 'form-group' },
        el('label', {}, 'Type'),
        el('select', { id: 'gift-type' },
          el('option', { value: 'given' }, 'Given'),
          el('option', { value: 'received' }, 'Received'),
        ),
      ),
    ),
    el('div', { class: 'form-group' },
      el('label', {}, 'Notes'),
      el('input', { type: 'text', id: 'gift-notes', placeholder: 'Notes' }),
    ),
    el('button', { onclick: () => addGift(name) }, 'Log gift'),
  );
}

function buildPrefsForm(name, existing) {
  return el('div', { class: 'add-form', id: 'prefs-form', style: 'display:none;' },
    el('h3', {}, 'Preferences'),
    el('div', { class: 'form-group' },
      el('textarea', { id: 'prefs-text', rows: 3, placeholder: 'Favorite colors, sizes, etc.' },
        existing || ''),
    ),
    el('button', { onclick: () => savePrefs(name) }, 'Save preferences'),
  );
}

// --- Tab content ---
function showTab(container, tab, contact, clickedTab) {
  container.innerHTML = '';
  // Update tab active states
  document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
  if (clickedTab) {
    clickedTab.classList.add('active');
  } else {
    // Default to first tab
    const first = document.querySelector('.tab');
    if (first) first.classList.add('active');
  }

  if (tab === 'ideas') {
    if (!contact.ideas || contact.ideas.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, el('p', {}, 'No ideas yet.')));
      return;
    }
    for (const idea of contact.ideas) {
      container.appendChild(renderIdeaCard(idea));
    }
  } else if (tab === 'given') {
    if (!contact.gifts_given || contact.gifts_given.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, el('p', {}, 'No gifts given yet.')));
      return;
    }
    for (const gift of contact.gifts_given) {
      container.appendChild(renderGiftCard(gift));
    }
  } else if (tab === 'received') {
    if (!contact.gifts_received || contact.gifts_received.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, el('p', {}, 'No gifts received yet.')));
      return;
    }
    for (const gift of contact.gifts_received) {
      container.appendChild(renderGiftCard(gift));
    }
  }
}

function renderIdeaCard(idea) {
  const card = el('div', { class: 'idea-item' });
  card.appendChild(el('div', { class: 'item-name' }, idea.item));

  const meta = el('div', { class: 'meta' });
  meta.appendChild(statusBadge(idea.status));
  if (idea.added) meta.appendChild(el('span', {}, `added: ${idea.added}`));
  if (idea.price_estimate) meta.appendChild(el('span', {}, `est: ${idea.currency || 'USD'} ${idea.price_estimate}`));
  if (idea.url) meta.appendChild(el('a', { href: idea.url, target: '_blank' }, 'link'));
  card.appendChild(meta);

  if (idea.tags && idea.tags.length > 0) {
    card.appendChild(el('div', { class: 'tags', style: 'margin-top: 6px;' }, ...idea.tags.map(tagPill)));
  }

  if (idea.notes) {
    card.appendChild(el('div', { class: 'text-muted', style: 'margin-top: 4px;' }, idea.notes));
  }

  card.appendChild(el('div', { class: 'id-line', style: 'margin-top: 4px;' }, idea.id));

  // Delete button
  const actions = el('div', { class: 'actions' });
  actions.appendChild(el('button', {
    class: 'danger btn-sm',
    onclick: () => deleteIdea(idea.id),
  }, 'Delete'));
  card.appendChild(actions);

  return card;
}

function renderGiftCard(gift) {
  const card = el('div', { class: 'gift-item' });
  card.appendChild(el('div', { class: 'item-name' }, gift.item));

  const meta = el('div', { class: 'meta' });
  if (gift.date) meta.appendChild(el('span', {}, `date: ${gift.date}`));
  if (gift.occasion) meta.appendChild(el('span', {}, `occasion: ${gift.occasion}`));
  if (gift.price) meta.appendChild(el('span', {}, `price: ${gift.currency || 'USD'} ${gift.price}`));
  card.appendChild(meta);

  if (gift.notes) {
    card.appendChild(el('div', { class: 'text-muted', style: 'margin-top: 4px;' }, gift.notes));
  }

  card.appendChild(el('div', { class: 'id-line', style: 'margin-top: 4px;' }, gift.id));
  return card;
}

// --- General ideas page ---

async function renderGeneralIdeasPage() {
  const app = document.getElementById('app');
  app.innerHTML = '';

  app.appendChild(header('General Ideas'));

  // Add general idea form (behind toggle)
  const addBtn = el('button', {
    class: 'secondary',
    onclick: () => {
      const form = document.getElementById('gi-form');
      form.style.display = form.style.display === 'none' ? 'block' : 'none';
    },
  }, '+ Add general idea');
  app.appendChild(addBtn);

  const addForm = el('div', { class: 'add-form', id: 'gi-form', style: 'display:none;' },
    el('h3', {}, 'Add general idea'),
    el('div', { class: 'form-group' },
      el('label', {}, 'Item'),
      el('input', { type: 'text', id: 'gi-item', placeholder: 'Ceramic pour-over set' }),
    ),
    el('div', { class: 'form-row' },
      el('div', { class: 'form-group' },
        el('label', {}, 'Tags'),
        el('input', { type: 'text', id: 'gi-tags', placeholder: 'kitchen, handmade' }),
      ),
      el('div', { class: 'form-group' },
        el('label', {}, 'URL'),
        el('input', { type: 'text', id: 'gi-url', placeholder: 'https://...' }),
      ),
    ),
    el('button', { onclick: addGeneralIdea }, 'Add general idea'),
  );
  app.appendChild(addForm);

  // Ideas list
  const listDiv = el('div', { class: 'mt-2' });
  app.appendChild(listDiv);

  try {
    const ideas = await api('/api/ideas');
    if (ideas.length === 0) {
      listDiv.appendChild(el('div', { class: 'empty-state' },
        el('p', {}, 'No general ideas yet.'),
      ));
      return;
    }
    for (const idea of ideas) {
      listDiv.appendChild(renderIdeaCard(idea));
    }
  } catch (err) {
    listDiv.appendChild(el('div', { class: 'empty-state' }, el('p', {}, `Error: ${err.message}`)));
  }
}

// --- Actions ---

async function quickAddIdea() {
  const name = document.getElementById('qa-for').value.trim();
  const item = document.getElementById('qa-item').value.trim();
  const tags = document.getElementById('qa-tags').value.trim();
  if (!name || !item) { alert('Name and item are required'); return; }

  try {
    await api(`/api/contacts/${encodeURIComponent(name)}/ideas`, {
      method: 'POST',
      body: JSON.stringify({ item, tags: tags ? tags.split(',').map(t => t.trim()) : [] }),
    });
    document.getElementById('qa-for').value = '';
    document.getElementById('qa-item').value = '';
    document.getElementById('qa-tags').value = '';
    document.getElementById('quick-add-form').style.display = 'none';
    router();
  } catch (err) { alert(err.message); }
}

async function addIdea(name) {
  const item = document.getElementById('add-item').value.trim();
  const tags = document.getElementById('add-tags').value.trim();
  const url = document.getElementById('add-url').value.trim();
  const notes = document.getElementById('add-notes').value.trim();
  if (!item) { alert('Item is required'); return; }

  try {
    await api(`/api/contacts/${encodeURIComponent(name)}/ideas`, {
      method: 'POST',
      body: JSON.stringify({ item, tags: tags ? tags.split(',').map(t => t.trim()) : [], url, notes }),
    });
    document.getElementById('idea-form').style.display = 'none';
    renderContactPage(name);
  } catch (err) { alert(err.message); }
}

async function addGift(name) {
  const item = document.getElementById('gift-item').value.trim();
  const occasion = document.getElementById('gift-occasion').value.trim();
  const date = document.getElementById('gift-date').value.trim();
  const price = document.getElementById('gift-price').value.trim();
  const type = document.getElementById('gift-type').value;
  const notes = document.getElementById('gift-notes').value.trim();
  if (!item) { alert('Item is required'); return; }

  const body = { item, occasion, date, notes };
  if (price) body.price = parseFloat(price);

  try {
    await api(`/api/contacts/${encodeURIComponent(name)}/gifts/${type}`, {
      method: 'POST',
      body: JSON.stringify(body),
    });
    document.getElementById('gift-form').style.display = 'none';
    renderContactPage(name);
  } catch (err) { alert(err.message); }
}

async function addGeneralIdea() {
  const item = document.getElementById('gi-item').value.trim();
  const tags = document.getElementById('gi-tags').value.trim();
  const url = document.getElementById('gi-url').value.trim();
  if (!item) { alert('Item is required'); return; }

  try {
    await api('/api/ideas', {
      method: 'POST',
      body: JSON.stringify({ item, tags: tags ? tags.split(',').map(t => t.trim()) : [], url }),
    });
    document.getElementById('gi-form').style.display = 'none';
    renderGeneralIdeasPage();
  } catch (err) { alert(err.message); }
}

async function savePrefs(name) {
  const prefs = document.getElementById('prefs-text').value.trim();
  try {
    await api(`/api/contacts/${encodeURIComponent(name)}/preferences`, {
      method: 'PUT',
      body: JSON.stringify({ preferences: prefs }),
    });
    document.getElementById('prefs-form').style.display = 'none';
    renderContactPage(name);
  } catch (err) { alert(err.message); }
}

async function deleteIdea(id) {
  if (!confirm('Delete this idea?')) return;
  try {
    // Determine if we're on a contact page or general ideas page
    const parts = currentRoute.split('/').filter(Boolean);
    if (parts[0] === 'contacts' && parts[1]) {
      const name = decodeURIComponent(parts[1]);
      await api(`/api/contacts/${encodeURIComponent(name)}/ideas/${id}`, { method: 'DELETE' });
    } else {
      await api(`/api/ideas/${id}`, { method: 'DELETE' });
    }
    router();
  } catch (err) { alert(err.message); }
}

// --- Init ---
router();