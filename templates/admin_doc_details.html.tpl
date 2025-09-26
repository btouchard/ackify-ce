{{define "admin_doc_details"}}
<div class="space-y-6">
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <div class="flex items-center space-x-3 mb-2">
          <a href="/admin" class="text-slate-400 hover:text-slate-600">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
            </svg>
          </a>
          <h1 class="text-2xl font-bold text-slate-900">Document {{.DocID}}</h1>
        </div>
        <p class="text-slate-600">Détails des signatures</p>
      </div>
      <div class="text-right">
        <div class="text-sm text-slate-500">Total signatures</div>
        <div class="text-2xl font-bold text-primary-600">{{len .Signatures}}</div>
      </div>
    </div>

    {{if .Signatures}}
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Utilisateur
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Date de signature
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Service
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              ID Utilisateur
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-slate-200">
          {{range .Signatures}}
          <tr class="hover:bg-slate-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div class="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center">
                  <svg class="w-4 h-4 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"/>
                  </svg>
                </div>
                <div class="ml-3">
                  {{if .UserName}}
                  <div class="text-sm font-medium text-slate-900">{{.UserName}}</div>
                  {{end}}
                  <div class="text-sm text-slate-500">{{.UserEmail}}</div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm text-slate-900">{{.SignedAtUTC.Format "02/01/2006"}}</div>
              <div class="text-sm text-slate-500">{{.SignedAtUTC.Format "15:04:05"}}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              {{$serviceInfo := .GetServiceInfo}}
              {{if $serviceInfo}}
              <div class="flex items-center">
                <img src="{{$serviceInfo.IconURL}}" alt="{{$serviceInfo.Name}}" class="w-4 h-4 mr-2">
                <span class="text-sm text-slate-900">{{$serviceInfo.Name}}</span>
              </div>
              {{else}}
              <span class="text-sm text-slate-500">-</span>
              {{end}}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <code class="text-xs text-slate-600 bg-slate-100 px-2 py-1 rounded">{{.UserSub}}</code>
            </td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{else}}
    <div class="text-center py-12">
      <div class="w-16 h-16 mx-auto mb-4 bg-slate-100 rounded-full flex items-center justify-center">
        <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      </div>
      <h3 class="text-lg font-medium text-slate-900 mb-2">Aucune signature</h3>
      <p class="text-slate-600">Ce document n'a pas encore été signé.</p>
    </div>
    {{end}}
  </div>

  {{if .Signatures}}
  <!-- Vérification de l'intégrité de la chaîne -->
  {{if .ChainIntegrity}}
  {{if .ChainIntegrity.IsValid}}
  <div class="bg-green-50 border border-green-200 rounded-lg p-4">
    <div class="flex">
      <div class="flex-shrink-0">
        <svg class="h-5 w-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
        </svg>
      </div>
      <div class="ml-3">
        <p class="text-sm text-green-700">
          <strong>Chaîne de blocs intègre :</strong> {{.ChainIntegrity.ValidSigs}}/{{.ChainIntegrity.TotalSigs}} signatures valides
        </p>
      </div>
    </div>
  </div>
  {{else}}
  <div class="bg-red-50 border border-red-200 rounded-lg p-4">
    <div class="flex">
      <div class="flex-shrink-0">
        <svg class="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>
        </svg>
      </div>
      <div class="ml-3">
        <p class="text-sm text-red-700">
          <strong>Problème d'intégrité détecté :</strong> {{.ChainIntegrity.InvalidSigs}} signature(s) invalide(s)
        </p>
        {{if .ChainIntegrity.Errors}}
        <div class="mt-2">
          <p class="text-xs text-red-600 font-medium">Erreurs détectées :</p>
          <ul class="mt-1 text-xs text-red-600 list-disc list-inside">
            {{range .ChainIntegrity.Errors}}
            <li>{{.}}</li>
            {{end}}
          </ul>
        </div>
        {{end}}
      </div>
    </div>
  </div>
  {{end}}
  {{end}}
  {{end}}
</div>
{{end}}