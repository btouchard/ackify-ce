{{define "base"}}<!doctype html>
<html lang="fr">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>Ackify - Proof of Read</title>
{{if .DocID}}
<link rel="alternate" type="application/json+oembed" href="/oembed?url={{.BaseURL}}/sign?doc={{.DocID}}&format=json" title="Signataires du document {{.DocID}}" />
{{end}}
<script src="https://cdn.tailwindcss.com"></script>
<script>
  tailwind.config = {
    theme: {
      extend: {
        colors: {
          primary: { 50: '#eff6ff', 100: '#dbeafe', 500: '#3b82f6', 600: '#2563eb', 700: '#1d4ed8', 900: '#1e3a8a' },
          success: { 50: '#f0fdf4', 100: '#dcfce7', 500: '#22c55e', 600: '#16a34a', 700: '#15803d' },
          warning: { 50: '#fffbeb', 100: '#fef3c7', 500: '#f59e0b', 600: '#d97706' },
          danger: { 50: '#fef2f2', 100: '#fecaca', 500: '#ef4444', 600: '#dc2626' }
        }
      }
    }
  }
</script>
</head>
<body class="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
  <div class="min-h-screen flex flex-col">
    <header class="bg-white/80 backdrop-blur-sm border-b border-slate-200 sticky top-0 z-10">
      <div class="max-w-4xl mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <h1 class="text-xl font-bold text-slate-900">Ackify - Proof of Read</h1>
          </div>
          {{if .User}}
            <div class="flex items-center space-x-4">
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
              <a href="/logout" class="text-sm text-slate-500 hover:text-slate-700 underline">Déconnexion</a>
            </div>
          {{end}}
        </div>
      </div>
    </header>
    
    <main class="flex-1 py-8">
      <div class="max-w-4xl mx-auto px-6">
        {{if eq .TemplateName "sign"}}{{template "sign" .}}{{else if eq .TemplateName "signatures"}}{{template "signatures" .}}{{else}}{{template "index" .}}{{end}}
      </div>
    </main>
    
    <footer class="bg-white/50 backdrop-blur-sm border-t border-slate-200 py-6">
      <div class="max-w-4xl mx-auto px-6">
        <div class="text-center space-y-2">
          <p class="text-xs text-slate-400">
            Développé par 
            <a href="mailto:benjamin@kolapsis.com" class="text-primary-600 hover:text-primary-700 font-medium">Benjamin Touchard</a>
            <span class="mx-1">•</span>
            <span class="text-slate-400">@2025</span>
          </p>
        </div>
      </div>
    </footer>
  </div>
</body>
</html>{{end}}