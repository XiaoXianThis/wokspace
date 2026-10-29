(function () {
  'use strict';

  // Install commands point at this site when it is served over http(s), so
  // they work on GitHub Pages and on a custom domain alike. The HTML keeps a
  // raw GitHub URL as the fallback.
  var root = document.documentElement.getAttribute('data-root') || '.';
  if (/^https?:$/.test(location.protocol)) {
    document.querySelectorAll('[data-install]').forEach(function (el) {
      var kind = el.getAttribute('data-install');
      var url = new URL(root + '/install.' + kind, location.href).href;
      el.textContent = kind === 'sh' ? 'curl -fsSL ' + url + ' | sh' : 'irm ' + url + ' | iex';
    });
  }

  // Tabs: pick Windows by default on Windows.
  document.querySelectorAll('.tabs').forEach(function (tabs) {
    var buttons = tabs.querySelectorAll('button[data-tab]');
    var group = tabs.parentElement;
    function select(name) {
      buttons.forEach(function (b) {
        b.setAttribute('aria-selected', String(b.getAttribute('data-tab') === name));
      });
      group.querySelectorAll('[data-panel]').forEach(function (p) {
        p.hidden = p.getAttribute('data-panel') !== name;
      });
    }
    buttons.forEach(function (b) {
      b.addEventListener('click', function () { select(b.getAttribute('data-tab')); });
    });
    var isWindows = /Windows/i.test(navigator.userAgent);
    select(isWindows ? 'windows' : buttons[0].getAttribute('data-tab'));
  });

  // Copy buttons.
  document.querySelectorAll('.code').forEach(function (block) {
    var pre = block.querySelector('pre');
    if (!pre || !navigator.clipboard) return;
    var button = document.createElement('button');
    button.className = 'copy';
    button.type = 'button';
    var label = document.documentElement.lang.indexOf('zh') === 0 ? ['复制', '已复制'] : ['Copy', 'Copied'];
    button.textContent = label[0];
    button.addEventListener('click', function () {
      var text = pre.innerText.split('\n').filter(function (line) {
        return line.trim() !== '' && line.trim().charAt(0) !== '#';
      }).map(function (line) {
        return line.replace(/\s+#\s.*$/, '');
      }).join('\n');
      navigator.clipboard.writeText(text).then(function () {
        button.textContent = label[1];
        setTimeout(function () { button.textContent = label[0]; }, 1500);
      });
    });
    block.appendChild(button);
  });
})();
