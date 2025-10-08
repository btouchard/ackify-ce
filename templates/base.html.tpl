{{define "base"}}<!doctype html>
<html lang="{{.Lang}}">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>{{index .T "site.title"}}</title>
{{if and (ne .TemplateName "admin_dashboard") (ne .TemplateName "admin_doc_details") (ne .TemplateName "admin_document_expected_signers")}}{{if .DocID}}
<link rel="alternate" type="application/json+oembed" href="/oembed?url={{.BaseURL}}/sign?doc={{.DocID}}&format=json" title="Signataires du document {{.DocID}}" />
{{end}}{{end}}
<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
<link rel="stylesheet" href="/static/output.css">
</head>
<body class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
  <div class="min-h-screen flex flex-col">
    <header class="bg-white/80 backdrop-blur-sm border-b border-slate-200 sticky top-0 z-10">
      <div class="max-w-4xl mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <a href="/" class="text-slate-400 hover:text-slate-600">
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
              <h1 class="text-xl font-bold text-slate-900">{{index .T "site.brand"}}</h1>
            </div>
          </a>

          <div class="flex items-center space-x-4">
            <!-- Language Switcher -->
            <div class="flex items-center space-x-1">
              <a href="#" onclick="switchLang('fr'); return false;" class="inline-flex items-center justify-center w-8 h-8 rounded-lg text-xl transition-all {{if eq .Lang "fr"}}bg-primary-50 ring-2 ring-primary-500{{else}}hover:bg-slate-100{{end}}" title="Français">
                🇫🇷
              </a>
              <a href="#" onclick="switchLang('en'); return false;" class="inline-flex items-center justify-center w-8 h-8 rounded-lg text-xl transition-all {{if eq .Lang "en"}}bg-primary-50 ring-2 ring-primary-500{{else}}hover:bg-slate-100{{end}}" title="English">
                🇬🇧
              </a>
            </div>
            <script>
              function switchLang(lang) {
                var currentPath = window.location.pathname + window.location.search;
                window.location.href = '/lang/' + lang + '?redirect=' + encodeURIComponent(currentPath);
              }
            </script>
            {{if .User}}
              <a href="/logout" onclick="localStorage.setItem('ackify_silent_login_attempted', Date.now().toString());" class="inline-flex items-center space-x-2 px-4 py-2 bg-white border border-slate-300 hover:border-red-500 hover:bg-red-50 text-slate-700 hover:text-red-700 text-sm font-medium rounded-xl transition-all duration-200 shadow-sm hover:shadow">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
                </svg>
                <span>{{if .User.Name}}{{.User.Name}}{{else}}{{.User.Email}}{{end}}</span>
              </a>
            {{else}}
              <a href="/login" class="inline-flex items-center space-x-2 px-4 py-2 bg-white border border-slate-300 hover:border-primary-500 hover:bg-primary-50 text-slate-700 hover:text-primary-700 text-sm font-medium rounded-xl transition-all duration-200 shadow-sm hover:shadow">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
                </svg>
                <span>{{index .T "header.login"}}</span>
              </a>
            {{end}}
          </div>
        </div>
      </div>

      <!-- Navigation Menu (visible when logged in) -->
      {{if .User}}
      <div class="border-t border-slate-200 bg-white/60">
        <div class="max-w-4xl mx-auto px-6">
          <nav class="flex items-center space-x-1 py-2">
            <a href="/" class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-slate-700 hover:text-primary-700 hover:bg-primary-50 rounded-lg transition-all">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
              </svg>
              <span>{{index .T "nav.home"}}</span>
            </a>
            <a href="/signatures" class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-slate-700 hover:text-primary-700 hover:bg-primary-50 rounded-lg transition-all">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <span>{{index .T "nav.signatures"}}</span>
            </a>
            {{if .IsAdmin}}
            <a href="/admin" class="inline-flex items-center space-x-2 px-4 py-2 text-sm font-medium text-slate-700 hover:text-primary-700 hover:bg-primary-50 rounded-lg transition-all">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
              <span>{{index .T "nav.admin"}}</span>
            </a>
            {{end}}
          </nav>
        </div>
      </div>
      {{end}}
    </header>

    <main class="flex-1 py-8">
      <div class="max-w-4xl mx-auto px-6">
        {{if eq .TemplateName "sign"}}
          {{template "sign" .}}
        {{else if eq .TemplateName "signatures"}}
          {{template "signatures" .}}
        {{else if eq .TemplateName "admin_dashboard"}}
          {{template "admin_dashboard" .}}
        {{else if eq .TemplateName "admin_doc_details"}}
          {{template "admin_doc_details" .}}
        {{else if eq .TemplateName "admin_document_expected_signers"}}
          {{template "admin_document_expected_signers" .}}
        {{else if eq .TemplateName "error"}}
          {{template "error" .}}
        {{else}}
          {{template "index" .}}
        {{end}}
      </div>
    </main>
    
    <footer class="bg-white/50 backdrop-blur-sm border-t border-slate-200 py-6">
      <div class="max-w-4xl mx-auto px-6">
        <div class="text-center space-y-2">
          <p class="text-xs text-slate-400">
            {{index .T "footer.developed_by"}}
            <a href="mailto:benjamin@kolapsis.com" class="text-primary-600 hover:text-primary-700 font-medium">Benjamin Touchard</a>
            <span class="mx-1">•</span>
            <span class="text-slate-400">{{index .T "footer.year"}}</span>
          </p>
        </div>
      </div>
    </footer>
  </div>

  {{if and (not .User) .AutoLogin}}
  <script>
    (function() {
      // Silent login: tente une connexion automatique si session OAuth existe
      const STORAGE_KEY = 'ackify_silent_login_attempted';
      const ATTEMPT_EXPIRY_MS = 5 * 60 * 1000; // 5 minutes

      function shouldAttemptSilentLogin() {
        const lastAttempt = localStorage.getItem(STORAGE_KEY);
        if (!lastAttempt) return true;

        const elapsed = Date.now() - parseInt(lastAttempt, 10);
        return elapsed > ATTEMPT_EXPIRY_MS;
      }

      function markSilentLoginAttempted() {
        localStorage.setItem(STORAGE_KEY, Date.now().toString());
      }

      function attemptSilentLogin() {
        if (!shouldAttemptSilentLogin()) {
          console.debug('[Silent Login] Tentative récente détectée, abandon');
          return;
        }

        console.debug('[Silent Login] Tentative de connexion silencieuse...');
        markSilentLoginAttempted();

        const currentURL = window.location.href;
        const loginURL = '/login?silent=true&next=' + encodeURIComponent(currentURL);

        // Redirection vers le provider OAuth avec prompt=none
        window.location.href = loginURL;
      }

      // Lancer la tentative de silent login au chargement
      if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', attemptSilentLogin);
      } else {
        attemptSilentLogin();
      }
    })();
  </script>
  {{end}}
</body>
</html>{{end}}