{{define "sign"}}
  <div class="space-y-6">
    <!-- Document Info Card -->
    <div class="bg-white rounded-3xl shadow-xl border border-slate-200 overflow-hidden">
      <div class="bg-gradient-to-r from-slate-100 to-slate-200 px-8 py-6 border-b border-slate-200">
        <div class="flex items-center space-x-4">
          <div class="w-10 h-10 bg-success-100 rounded-xl flex items-center justify-center flex-shrink-0">
            <svg class="w-5 h-5 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-3">
              <p class="font-semibold text-slate-900 text-2xl">Document {{.DocID}}</p>
              {{if .ServiceInfo}}
                <div class="flex items-center space-x-1 bg-slate-100 px-2 py-1 rounded-md">
                  <img src="{{.ServiceInfo.Icon}}" alt="{{.ServiceInfo.Name}}" class="w-3 h-3">
                  <span class="text-xs text-slate-600">{{.ServiceInfo.Name}}</span>
                </div>
              {{end}}
            </div>
          </div>
        </div>
      </div>
      
      <div class="px-8 py-8">
        {{if .Already}}
          <!-- Document Already Signed -->
          <div class="text-center space-y-6">
            <div class="mx-auto w-20 h-20 bg-success-100 rounded-full flex items-center justify-center">
              <svg class="w-10 h-10 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            
            <div>
              <h3 class="text-xl font-bold text-success-700 mb-2">Document signé</h3>
              <p class="text-slate-600 mb-4">Vous avez confirmé la lecture de ce document</p>
              
              <div class="bg-success-50 border border-success-200 rounded-2xl p-6">
                <div class="flex items-center justify-center space-x-3 text-success-800">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                  <span class="font-semibold">Signé le {{.SignedAt}}</span>
                </div>
                <p class="text-success-700 text-sm mt-2">Signature cryptographique enregistrée et vérifiable</p>
              </div>
            </div>
          </div>
        {{else}}
          <!-- Document Not Signed Yet -->
          <div class="text-center space-y-6">
            <div class="mx-auto w-20 h-20 bg-warning-100 rounded-full flex items-center justify-center">
              <svg class="w-10 h-10 text-warning-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"/>
              </svg>
            </div>
            
            <div>
              <h3 class="text-xl font-bold text-warning-700 mb-2">Document à signer</h3>
              <p class="text-slate-600 mb-6">Vous devez confirmer avoir lu et compris ce document</p>
              
              <div class="bg-warning-50 border border-warning-200 rounded-2xl p-6 mb-6">
                <div class="flex items-start space-x-3">
                  <svg class="w-5 h-5 text-warning-600 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                  <div class="text-left">
                    <p class="font-semibold text-warning-800 mb-1">Avant de signer</p>
                    <p class="text-warning-700 text-sm">Assurez-vous d'avoir lu et compris l'intégralité du document. La signature est irréversible.</p>
                  </div>
                </div>
              </div>
              
              <form method="POST" action="/sign">
                <input type="hidden" name="doc" value="{{.DocID}}" />
                {{if .ServiceInfo}}
                  <input type="hidden" name="referrer" value="{{.ServiceInfo.Referrer}}" />
                {{end}}
                <button type="submit" class="w-full bg-gradient-to-r from-success-600 to-success-700 hover:from-success-700 hover:to-success-800 text-white font-bold py-4 px-8 rounded-2xl transition-all duration-200 flex items-center justify-center space-x-3 shadow-lg hover:shadow-xl text-lg">
                  <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
                  </svg>
                  <span>Je certifie avoir lu et compris ce document</span>
                </button>
              </form>
            </div>
          </div>
        {{end}}
      </div>
    </div>
    
    <!-- Actions Card -->
    <div class="bg-white rounded-2xl border border-slate-200 p-6">
      <h4 class="font-semibold text-slate-900 mb-4">Actions supplémentaires</h4>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <a href="/signatures" target="_blank" class="flex items-center justify-center space-x-2 px-4 py-3 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl transition-colors text-sm font-medium text-center">
          <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <span>Voir mes signatures</span>
        </a>
        
        <a href="/" class="flex items-center justify-center space-x-2 px-4 py-3 bg-primary-100 hover:bg-primary-200 text-primary-700 rounded-xl transition-colors text-sm font-medium text-center">
          <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"/>
          </svg>
          <span>Retour à l'accueil</span>
        </a>
      </div>
    </div>
  </div>
{{end}}