{{define "error"}}
  <div class="min-h-[60vh] flex items-center justify-center">
    <div class="max-w-2xl w-full mx-4">
      <div class="bg-white rounded-3xl shadow-2xl border border-slate-200 overflow-hidden">
        <div class="bg-gradient-to-r from-red-600 to-red-700 px-8 py-6">
          <div class="flex items-center space-x-4">
            <div class="w-12 h-12 bg-white/20 rounded-2xl flex items-center justify-center">
              <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
            </div>
            <div>
              <h2 class="text-2xl font-bold text-white">{{.ErrorTitle}}</h2>
            </div>
          </div>
        </div>

        <div class="px-8 py-12 text-center">
          <p class="text-lg text-slate-700 mb-8">
            {{.ErrorMessage}}
          </p>

          {{if .User}}
          <div class="bg-slate-50 border border-slate-200 rounded-2xl p-6 mb-8">
            <p class="text-sm text-slate-600 mb-2">{{index .T "error.connected_as"}}</p>
            <p class="font-semibold text-slate-900">{{.User.Email}}</p>
          </div>
          {{end}}

          <div class="flex justify-center space-x-4">
            <a href="/" class="inline-flex items-center px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white font-semibold rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl">
              <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
              </svg>
              {{index .T "error.back_home"}}
            </a>

            {{if .User}}
            <a href="/logout" class="inline-flex items-center px-6 py-3 bg-slate-600 hover:bg-slate-700 text-white font-semibold rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl">
              <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
              </svg>
              {{index .T "error.sign_out"}}
            </a>
            {{end}}
          </div>
        </div>
      </div>
    </div>
  </div>
{{end}}