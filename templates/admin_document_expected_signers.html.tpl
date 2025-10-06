{{define "admin_document_expected_signers"}}
<div class="space-y-6">
  <!-- Header -->
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
        <p class="text-slate-600">{{if eq .Lang "fr"}}Gestion des signataires attendus{{else}}Expected Signers Management{{end}}</p>
      </div>
    </div>

    <!-- Stats Cards -->
    {{if gt .Stats.ExpectedCount 0}}
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div class="text-sm text-blue-600 font-medium">{{if eq .Lang "fr"}}Attendus{{else}}Expected{{end}}</div>
        <div class="text-2xl font-bold text-blue-900">{{.Stats.ExpectedCount}}</div>
      </div>
      <div class="bg-green-50 border border-green-200 rounded-lg p-4">
        <div class="text-sm text-green-600 font-medium">{{if eq .Lang "fr"}}Signés{{else}}Signed{{end}}</div>
        <div class="text-2xl font-bold text-green-900">{{.Stats.SignedCount}}</div>
      </div>
      <div class="bg-orange-50 border border-orange-200 rounded-lg p-4">
        <div class="text-sm text-orange-600 font-medium">{{if eq .Lang "fr"}}En attente{{else}}Pending{{end}}</div>
        <div class="text-2xl font-bold text-orange-900">{{.Stats.PendingCount}}</div>
      </div>
      <div class="bg-purple-50 border border-purple-200 rounded-lg p-4">
        <div class="text-sm text-purple-600 font-medium">{{if eq .Lang "fr"}}Taux de complétion{{else}}Completion Rate{{end}}</div>
        <div class="text-2xl font-bold text-purple-900">{{printf "%.0f" .Stats.CompletionRate}}%</div>
      </div>
    </div>

    <!-- Progress Bar -->
    <div class="mb-6">
      <div class="flex justify-between text-sm text-slate-600 mb-2">
        <span>{{if eq .Lang "fr"}}Progression{{else}}Progress{{end}}</span>
        <span>{{.Stats.SignedCount}} / {{.Stats.ExpectedCount}}</span>
      </div>
      <div class="w-full bg-slate-200 rounded-full h-3 overflow-hidden">
        <div class="bg-gradient-to-r from-blue-500 to-purple-600 h-3 rounded-full transition-all duration-500" style="width: {{printf "%.0f" .Stats.CompletionRate}}%"></div>
      </div>
    </div>
    {{end}}

    <!-- Share Link Section -->
    <div class="bg-slate-50 border border-slate-200 rounded-lg p-4">
      <div class="text-sm font-medium text-slate-700 mb-2">
        {{if eq .Lang "fr"}}Lien à partager{{else}}Share Link{{end}}
      </div>
      <div class="flex items-center space-x-2">
        <input type="text" value="{{.ShareLink}}" readonly class="flex-1 px-3 py-2 text-sm border border-slate-300 rounded-lg bg-white font-mono" id="shareLink">
        <button onclick="copyShareLink()" class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors">
          {{if eq .Lang "fr"}}Copier{{else}}Copy{{end}}
        </button>
      </div>
      <div id="copyFeedback" class="hidden mt-2 text-sm text-green-600">
        {{if eq .Lang "fr"}}✓ Lien copié !{{else}}✓ Link copied!{{end}}
      </div>
    </div>
  </div>

  <!-- Add Expected Signers Form -->
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <h2 class="text-lg font-semibold text-slate-900 mb-4">
      {{if eq .Lang "fr"}}Ajouter des signataires attendus{{else}}Add Expected Signers{{end}}
    </h2>
    <form method="POST" action="/admin/docs/{{.DocID}}/expected">
      <div class="mb-4">
        <label for="emails" class="block text-sm font-medium text-slate-700 mb-2">
          {{if eq .Lang "fr"}}Emails (un par ligne, ou séparés par virgule/point-virgule){{else}}Emails (one per line, or separated by comma/semicolon){{end}}
        </label>
        <textarea name="emails" id="emails" rows="5" class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent" placeholder="user1@example.com&#10;user2@example.com&#10;user3@example.com"></textarea>
      </div>
      <button type="submit" class="px-4 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors">
        {{if eq .Lang "fr"}}Ajouter{{else}}Add{{end}}
      </button>
    </form>
  </div>

  <!-- Expected Signers Table -->
  {{if .ExpectedSigners}}
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <h2 class="text-lg font-semibold text-slate-900 mb-4">
      {{if eq .Lang "fr"}}Liste des signataires attendus{{else}}Expected Signers List{{end}}
    </h2>
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Email
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Statut{{else}}Status{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Signé le{{else}}Signed At{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-slate-200">
          {{range .ExpectedSigners}}
          <tr class="hover:bg-slate-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-slate-900">{{.Email}}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              {{if .HasSigned}}
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                ✓ {{if eq $.Lang "fr"}}Signé{{else}}Signed{{end}}
              </span>
              {{else}}
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
                ⏳ {{if eq $.Lang "fr"}}En attente{{else}}Pending{{end}}
              </span>
              {{end}}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
              {{if .SignedAt}}
              <div>{{.SignedAt.Format "02/01/2006"}}</div>
              <div class="text-xs">{{.SignedAt.Format "15:04:05"}}</div>
              {{else}}
              -
              {{end}}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm">
              <form method="POST" action="/admin/docs/{{$.DocID}}/expected/remove" class="inline">
                <input type="hidden" name="email" value="{{.Email}}">
                <button type="submit" class="text-red-600 hover:text-red-900 font-medium" onclick="return confirm('{{if eq $.Lang "fr"}}Supprimer ce signataire attendu ?{{else}}Remove this expected signer?{{end}}')">
                  {{if eq $.Lang "fr"}}Retirer{{else}}Remove{{end}}
                </button>
              </form>
            </td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
  {{end}}

  <!-- Unexpected Signatures Section -->
  {{if .UnexpectedSignatures}}
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <h2 class="text-lg font-semibold text-slate-900 mb-4">
      {{if eq .Lang "fr"}}Signatures non attendues{{else}}Unexpected Signatures{{end}}
    </h2>
    <p class="text-sm text-slate-600 mb-4">
      {{if eq .Lang "fr"}}Ces utilisateurs ont signé mais n'étaient pas dans la liste des signataires attendus.{{else}}These users signed but were not in the expected signers list.{{end}}
    </p>
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Email
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Signé le{{else}}Signed At{{end}}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-slate-200">
          {{range .UnexpectedSignatures}}
          <tr class="hover:bg-slate-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-slate-900">{{.UserEmail}}</div>
              {{if .UserName}}
              <div class="text-sm text-slate-500">{{.UserName}}</div>
              {{end}}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-500">
              <div>{{.SignedAtUTC.Format "02/01/2006"}}</div>
              <div class="text-xs">{{.SignedAtUTC.Format "15:04:05"}}</div>
            </td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
  {{end}}

  <!-- All Signatures Section -->
  {{if .Signatures}}
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <h2 class="text-lg font-semibold text-slate-900 mb-4">
      {{if eq .Lang "fr"}}Toutes les signatures{{else}}All Signatures{{end}} ({{len .Signatures}})
    </h2>
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Utilisateur{{else}}User{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Signé le{{else}}Signed At{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Service
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
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>

  <!-- Chain Integrity Section -->
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
          <strong>{{if eq .Lang "fr"}}Intégrité de la chaîne validée{{else}}Chain integrity valid{{end}}</strong> - {{.ChainIntegrity.ValidSigs}}/{{.ChainIntegrity.TotalSigs}} signatures
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
          <strong>{{if eq .Lang "fr"}}Problème d'intégrité détecté{{else}}Chain integrity issues{{end}}</strong> - {{.ChainIntegrity.InvalidSigs}} {{if eq .Lang "fr"}}erreurs{{else}}errors{{end}}
        </p>
        {{if .ChainIntegrity.Errors}}
        <div class="mt-2">
          <p class="text-xs text-red-600 font-medium">{{if eq .Lang "fr"}}Erreurs détectées :{{else}}Detected errors:{{end}}</p>
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
  {{else}}
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="text-center py-12">
      <div class="w-16 h-16 mx-auto mb-4 bg-slate-100 rounded-full flex items-center justify-center">
        <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
      </div>
      <h3 class="text-lg font-medium text-slate-900 mb-2">
        {{if eq .Lang "fr"}}Aucune signature pour le moment{{else}}No signatures yet{{end}}
      </h3>
      <p class="text-slate-600">
        {{if eq .Lang "fr"}}Ce document n'a pas encore été signé.{{else}}This document has not been signed yet.{{end}}
      </p>
    </div>
  </div>
  {{end}}
</div>

<script>
function copyShareLink() {
  const linkInput = document.getElementById('shareLink');
  const feedback = document.getElementById('copyFeedback');

  linkInput.select();
  linkInput.setSelectionRange(0, 99999); // For mobile devices

  try {
    document.execCommand('copy');
    feedback.classList.remove('hidden');
    setTimeout(() => {
      feedback.classList.add('hidden');
    }, 3000);
  } catch (err) {
    console.error('Failed to copy:', err);
  }
}
</script>
{{end}}
