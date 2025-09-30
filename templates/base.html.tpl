{{define "base"}}<!doctype html>
<html lang="{{.Lang}}">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>{{index .T "site.title"}}</title>
{{if and (ne .TemplateName "admin_dashboard") (ne .TemplateName "admin_doc_details")}}{{if .DocID}}
<link rel="alternate" type="application/json+oembed" href="/oembed?url={{.BaseURL}}/sign?doc={{.DocID}}&format=json" title="Signataires du document {{.DocID}}" />
{{end}}{{end}}
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
              <div class="text-sm text-slate-600">
                <span class="inline-flex items-center space-x-2">
                  <div class="w-6 h-6 bg-primary-100 rounded-full flex items-center justify-center">
                    <svg class="w-3 h-3 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"/>
                    </svg>
                  </div>
                  <span>{{if .User.Name}}{{.User.Name}}{{else}}{{.User.Email}}{{end}}</span>
                </span>
              </div>
              <a href="/logout" class="text-sm text-slate-500 hover:text-slate-700 underline">{{index .T "header.logout"}}</a>
            {{end}}
          </div>
        </div>
      </div>
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
</body>
</html>{{end}}