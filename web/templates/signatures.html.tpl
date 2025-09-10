{{define "signatures"}}
  <div class="space-y-6">
    <!-- Header -->
    <div class="bg-white rounded-3xl shadow-xl border border-slate-200 overflow-hidden">
      <div class="bg-gradient-to-r from-primary-600 to-primary-700 px-8 py-6">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <div class="w-12 h-12 bg-white/20 rounded-2xl flex items-center justify-center">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <div>
              <h2 class="text-2xl font-bold text-white">Mes signatures</h2>
              <p class="text-primary-100">Liste de tous les documents que vous avez signés</p>
            </div>
          </div>
          <a href="/" class="text-primary-100 hover:text-white transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
            </svg>
          </a>
        </div>
      </div>
    </div>

    <!-- Signatures List -->
    <div class="bg-white rounded-3xl shadow-xl border border-slate-200 overflow-hidden">
      {{if .Signatures}}
        <!-- Stats -->
        <div class="bg-gradient-to-r from-slate-50 to-slate-100 px-8 py-4 border-b border-slate-200">
          <div class="flex items-center justify-between">
            <span class="text-slate-600 font-medium">
              {{len .Signatures}} signature{{if gt (len .Signatures) 1}}s{{end}} au total
            </span>
            <span class="text-sm text-slate-500">
              Trié par date décroissante
            </span>
          </div>
        </div>

        <!-- Signatures -->
        <div class="divide-y divide-slate-100">
          {{range .Signatures}}
            <div class="px-8 py-6 hover:bg-slate-50 transition-colors">
              <div class="flex items-center space-x-4">
                <div class="w-10 h-10 bg-success-100 rounded-xl flex items-center justify-center flex-shrink-0">
                  <svg class="w-5 h-5 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="flex items-center space-x-3">
                        <p class="font-semibold text-slate-900">Document {{.DocID}}</p>
                        {{if .GetServiceInfo}}
                          <div class="flex items-center space-x-1 bg-slate-100 px-2 py-1 rounded-md">
                            <img src="{{.GetServiceInfo.Icon}}" alt="{{.GetServiceInfo.Name}}" class="w-3 h-3">
                            <span class="text-xs text-slate-600">{{.GetServiceInfo.Name}}</span>
                          </div>
                        {{end}}
                      </div>
                      <p class="text-sm text-slate-500 mt-1">
                        Signé le {{.SignedAtUTC.Format "02/01/2006 à 15:04:05"}}
                      </p>
                    </div>
                    <div class="flex space-x-2">
                      <a href="/sign?doc={{.DocID}}" 
                         class="inline-flex items-center px-3 py-2 text-sm font-medium text-primary-700 bg-primary-50 rounded-lg hover:bg-primary-100 transition-colors">
                        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                        </svg>
                        Voir
                      </a>
                      <a href="/status?doc={{.DocID}}" 
                         class="inline-flex items-center px-3 py-2 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 transition-colors">
                        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                        </svg>
                        Statut
                      </a>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {{end}}
        </div>
      {{else}}
        <!-- Empty State -->
        <div class="px-8 py-16 text-center">
          <div class="w-20 h-20 mx-auto bg-slate-100 rounded-2xl flex items-center justify-center mb-6">
            <svg class="w-10 h-10 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
            </svg>
          </div>
          <h3 class="text-lg font-semibold text-slate-900 mb-2">Aucune signature</h3>
          <p class="text-slate-500 mb-6">Vous n'avez encore signé aucun document.</p>
          <a href="/" 
             class="inline-flex items-center px-6 py-3 bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-700 hover:to-primary-800 text-white font-semibold rounded-2xl transition-all duration-200 shadow-lg hover:shadow-xl">
            <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
            </svg>
            Signer un document
          </a>
        </div>
      {{end}}
    </div>
  </div>
{{end}}