{{define "index"}}
  <div class="space-y-8">
    <!-- Hero Section -->
    <div class="bg-white rounded-3xl shadow-xl border border-slate-200 overflow-hidden">
      <div class="bg-gradient-to-r from-primary-600 to-primary-700 px-8 py-6">
        <div class="flex items-center space-x-4">
          <div class="w-12 h-12 bg-white/20 rounded-2xl flex items-center justify-center">
            <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
            </svg>
          </div>
          <div>
            <h2 class="text-2xl font-bold text-white">{{index .T "home.title"}}</h2>
            <p class="text-primary-100">{{index .T "home.subtitle"}}</p>
          </div>
        </div>
      </div>
      
      <div class="px-8 py-8">
        <form method="GET" action="/sign" class="space-y-6">
          <div>
            <label for="doc" class="block text-sm font-semibold text-slate-700 mb-3">
              {{index .T "home.doc_label"}}
            </label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                <svg class="h-5 w-5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a.997.997 0 01-.707.293H7a4 4 0 01-4-4V7a4 4 0 014-4z"/>
                </svg>
              </div>
              <input
                id="doc"
                name="doc"
                placeholder="{{index .T "home.doc_placeholder"}}"
                class="block w-full pl-12 pr-4 py-4 border border-slate-300 rounded-2xl text-lg placeholder-slate-400 focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                required
              />
            </div>
            <p class="mt-2 text-sm text-slate-500">{{index .T "home.doc_help"}}</p>
          </div>

          <button type="submit" class="w-full bg-gradient-to-r from-primary-600 to-primary-700 hover:from-primary-500 hover:to-primary-600 text-white font-semibold py-4 px-6 rounded-2xl transition-all duration-200 flex items-center justify-center space-x-3 shadow-lg hover:shadow-xl hover:scale-[1.02]">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6"/>
            </svg>
            <span>{{index .T "home.submit"}}</span>
          </button>
        </form>
      </div>
    </div>
    
    <!-- Info Cards -->
    <div class="grid md:grid-cols-3 gap-6">
      <div class="bg-white rounded-2xl p-6 border border-slate-200 hover:shadow-lg transition-shadow">
        <div class="w-10 h-10 bg-success-100 rounded-xl flex items-center justify-center mb-4">
          <svg class="w-5 h-5 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
          </svg>
        </div>
        <h3 class="font-semibold text-slate-900 mb-2">{{index .T "home.feature_secure_title"}}</h3>
        <p class="text-sm text-slate-600">{{index .T "home.feature_secure_desc"}}</p>
      </div>

      <div class="bg-white rounded-2xl p-6 border border-slate-200 hover:shadow-lg transition-shadow">
        <div class="w-10 h-10 bg-primary-100 rounded-xl flex items-center justify-center mb-4">
          <svg class="w-5 h-5 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
        </div>
        <h3 class="font-semibold text-slate-900 mb-2">{{index .T "home.feature_efficient_title"}}</h3>
        <p class="text-sm text-slate-600">{{index .T "home.feature_efficient_desc"}}</p>
      </div>

      <div class="bg-white rounded-2xl p-6 border border-slate-200 hover:shadow-lg transition-shadow">
        <div class="w-10 h-10 bg-warning-100 rounded-xl flex items-center justify-center mb-4">
          <svg class="w-5 h-5 text-warning-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
          </svg>
        </div>
        <h3 class="font-semibold text-slate-900 mb-2">{{index .T "home.feature_compliant_title"}}</h3>
        <p class="text-sm text-slate-600">{{index .T "home.feature_compliant_desc"}}</p>
      </div>
    </div>
  </div>
{{end}}