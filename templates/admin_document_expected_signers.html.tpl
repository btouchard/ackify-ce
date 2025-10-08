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
        <p class="text-slate-600">{{if eq .Lang "fr"}}Gestion des confirmations de lecture attendues{{else}}Expected Readers Management{{end}}</p>
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
        <div class="text-sm text-green-600 font-medium">{{if eq .Lang "fr"}}Confirmés{{else}}Confirmed{{end}}</div>
        <div class="flex items-baseline justify-between">
          <div class="text-2xl font-bold text-green-900">{{.Stats.SignedCount}}</div>
          {{if gt (len .UnexpectedSignatures) 0}}
          <div class="text-sm font-medium text-green-700">+{{len .UnexpectedSignatures}}</div>
          {{end}}
        </div>
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

  <!-- Document Metadata Section -->
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-slate-900">
        {{if eq .Lang "fr"}}📄 Métadonnées du document{{else}}📄 Document Metadata{{end}}
      </h2>
      <button onclick="openEditMetadataModal()" class="inline-flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
        </svg>
        <span>{{if eq .Lang "fr"}}Modifier{{else}}Edit{{end}}</span>
      </button>
    </div>

    {{if .Document}}
    <div class="space-y-4">
      {{if .Document.Title}}
      <div>
        <div class="text-sm font-medium text-slate-700 mb-1">{{if eq .Lang "fr"}}Titre{{else}}Title{{end}}</div>
        <div class="text-slate-900">{{.Document.Title}}</div>
      </div>
      {{end}}

      {{if .Document.URL}}
      <div>
        <div class="text-sm font-medium text-slate-700 mb-1">{{if eq .Lang "fr"}}URL / Emplacement{{else}}URL / Location{{end}}</div>
        <div class="text-slate-900">
          <a href="{{.Document.URL}}" target="_blank" rel="noopener noreferrer" class="text-primary-600 hover:text-primary-700 hover:underline inline-flex items-center space-x-1">
            <span class="break-all">{{.Document.URL}}</span>
            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
            </svg>
          </a>
        </div>
      </div>
      {{end}}

      {{if .Document.Checksum}}
      <div>
        <div class="text-sm font-medium text-slate-700 mb-1">
          {{if eq .Lang "fr"}}Empreinte ({{.Document.ChecksumAlgorithm}}){{else}}Checksum ({{.Document.ChecksumAlgorithm}}){{end}}
        </div>
        <div class="flex items-center space-x-2">
          <input type="text" value="{{.Document.Checksum}}" readonly class="flex-1 px-3 py-2 text-sm border border-slate-300 rounded-lg bg-slate-50 font-mono text-xs" id="docChecksum">
          <button onclick="copyChecksum()" class="px-4 py-2 bg-slate-600 text-white text-sm font-medium rounded-lg hover:bg-slate-700 transition-colors">
            {{if eq .Lang "fr"}}Copier{{else}}Copy{{end}}
          </button>
        </div>
        <div id="checksumCopyFeedback" class="hidden mt-2 text-sm text-green-600">
          {{if eq .Lang "fr"}}✓ Empreinte copiée !{{else}}✓ Checksum copied!{{end}}
        </div>
      </div>
      {{end}}

      {{if .Document.Description}}
      <div>
        <div class="text-sm font-medium text-slate-700 mb-1">{{if eq .Lang "fr"}}Description{{else}}Description{{end}}</div>
        <div class="text-slate-900 whitespace-pre-wrap">{{.Document.Description}}</div>
      </div>
      {{end}}

      <div class="text-xs text-slate-500 pt-2 border-t border-slate-200">
        {{if eq .Lang "fr"}}Créé par{{else}}Created by{{end}} {{.Document.CreatedBy}} {{if eq .Lang "fr"}}le{{else}}on{{end}} {{.Document.CreatedAt.Format "2006-01-02 15:04"}}
        {{if not (.Document.UpdatedAt.Equal .Document.CreatedAt)}}
        • {{if eq .Lang "fr"}}Modifié le{{else}}Updated on{{end}} {{.Document.UpdatedAt.Format "2006-01-02 15:04"}}
        {{end}}
      </div>
    </div>
    {{else}}
    <div class="text-center py-8 text-slate-500">
      <svg class="w-16 h-16 mx-auto mb-4 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
      </svg>
      <p class="text-sm">{{if eq .Lang "fr"}}Aucune métadonnée pour ce document{{else}}No metadata for this document{{end}}</p>
      <button onclick="openEditMetadataModal()" class="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium">
        {{if eq .Lang "fr"}}Ajouter des métadonnées{{else}}Add metadata{{end}}
      </button>
    </div>
    {{end}}
  </div>

  <!-- Expected Signers Table -->
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-slate-900">
        {{if eq .Lang "fr"}}✓ Confirmations de lecture attendues{{else}}✓ Expected Readers{{end}}
        {{if .ExpectedSigners}}
        <span class="text-sm font-medium text-slate-600 ml-2">
          ({{.Stats.SignedCount}}/{{.Stats.ExpectedCount}})
        </span>
        {{end}}
      </h2>
      <button onclick="openAddSignersModal()" class="inline-flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        <span>{{if eq .Lang "fr"}}Ajouter{{else}}Add{{end}}</span>
      </button>
    </div>

    {{if .ExpectedSigners}}
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Lecteur{{else}}Reader{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Confirmé le{{else}}Confirmed At{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-slate-200">
          {{range .ExpectedSigners}}
          <tr class="hover:bg-slate-50">
            <td class="px-6 py-4">
              <div class="flex items-center space-x-3">
                {{if not .HasSigned}}
                <input type="checkbox" name="emails" value="{{.Email}}" class="email-checkbox rounded text-primary-600 focus:ring-primary-500 h-4 w-4">
                {{else}}
                <div class="h-4 w-4"></div>
                {{end}}
                <div class="flex items-center space-x-2">
                  <div>
                    {{if .HasSigned}}
                    <span class="text-green-600">✓</span>
                    {{else}}
                    <span class="text-orange-500">⏳</span>
                    {{end}}
                  </div>
                  <div>
                    <div class="text-sm font-medium text-slate-900">
                      {{if and .UserName .HasSigned}}
                        {{.UserName}} &lt;{{.Email}}&gt;
                      {{else}}
                        {{if .Name}}
                          {{.Name}} &lt;{{.Email}}&gt;
                        {{else}}
                          {{.Email}}
                        {{end}}
                      {{end}}
                    </div>
                  </div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 text-sm text-slate-500">
              {{if .SignedAt}}
              <div>{{.SignedAt.Format "02/01 15:04"}}</div>
              {{else}}
              <span class="text-slate-400">{{if eq $.Lang "fr"}}En attente{{else}}Pending{{end}}</span>
              {{end}}
            </td>
            <td class="px-6 py-4 text-sm text-slate-500">
              {{if .LastReminderSent}}
              <div class="space-y-1">
                <div>{{.LastReminderSent.Format "02/01 15:04"}}</div>
                <div class="text-xs text-slate-400">
                  ({{.ReminderCount}} {{if eq $.Lang "fr"}}envoi(s){{else}}sent{{end}})
                </div>
              </div>
              {{else}}
              <span class="text-slate-400">{{if eq $.Lang "fr"}}Jamais{{else}}Never{{end}}</span>
              {{end}}
            </td>
            <td class="px-6 py-4 text-sm">
              <form method="POST" action="/admin/docs/{{$.DocID}}/expected/remove" class="inline" onsubmit="event.preventDefault(); showDeleteModal(this);">
                <input type="hidden" name="email" value="{{.Email}}">
                <button type="submit" class="text-red-600 hover:text-red-900 font-medium">
                  {{if eq $.Lang "fr"}}Retirer{{else}}Remove{{end}}
                </button>
              </form>
            </td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
    {{else}}
    <div class="text-center py-8 text-slate-500">
      <p>{{if eq .Lang "fr"}}Aucun lecteur attendu pour le moment{{else}}No expected readers yet{{end}}</p>
    </div>
    {{end}}
  </div>

  <!-- Email Reminders Section -->
  {{if and .ReminderStats (gt .Stats.ExpectedCount 0)}}
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-slate-900">
        {{if eq .Lang "fr"}}📧 Relances par email{{else}}📧 Email Reminders{{end}}
      </h2>
    </div>

    <!-- Reminder Stats -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div class="text-sm text-blue-600 font-medium">{{if eq .Lang "fr"}}Relances envoyées{{else}}Reminders Sent{{end}}</div>
        <div class="text-2xl font-bold text-blue-900">{{.ReminderStats.TotalSent}}</div>
      </div>
      <div class="bg-orange-50 border border-orange-200 rounded-lg p-4">
        <div class="text-sm text-orange-600 font-medium">{{if eq .Lang "fr"}}À relancer{{else}}To Remind{{end}}</div>
        <div class="text-2xl font-bold text-orange-900">{{.ReminderStats.PendingCount}}</div>
      </div>
      {{if .ReminderStats.LastSentAt}}
      <div class="bg-green-50 border border-green-200 rounded-lg p-4">
        <div class="text-sm text-green-600 font-medium">{{if eq .Lang "fr"}}Dernière relance{{else}}Last Reminder{{end}}</div>
        <div class="text-sm font-bold text-green-900">{{.ReminderStats.LastSentAt.Format "02/01 15:04"}}</div>
      </div>
      {{end}}
    </div>

    <!-- Send Reminders Form -->
    {{if gt .ReminderStats.PendingCount 0}}
    <form method="POST" action="/admin/docs/{{.DocID}}/reminders/send" onsubmit="return confirmSendReminders()">
      <div class="bg-slate-50 border border-slate-200 rounded-lg p-4">
        <div class="space-y-3">
          {{if .Document}}
          {{if .Document.URL}}
          <div class="text-sm text-slate-600 mb-2">
            <span class="font-medium">{{if eq .Lang "fr"}}Document :{{else}}Document:{{end}}</span>
            <a href="{{.Document.URL}}" target="_blank" rel="noopener noreferrer" class="text-primary-600 hover:text-primary-700 hover:underline ml-1">
              {{.Document.URL}}
            </a>
          </div>
          {{end}}
          {{end}}

          <div class="text-sm font-medium text-slate-700 mb-2">
            {{if eq .Lang "fr"}}Envoyer des relances :{{else}}Send reminders:{{end}}
          </div>

          <label class="flex items-center space-x-2 cursor-pointer">
            <input type="radio" name="send_mode" value="all" class="text-primary-600" checked>
            <span class="text-sm text-slate-700">
              {{if eq .Lang "fr"}}Envoyer à tous les lecteurs en attente{{else}}Send to all pending readers{{end}}
              <span class="font-semibold">({{.ReminderStats.PendingCount}})</span>
            </span>
          </label>

          <label class="flex items-center space-x-2 cursor-pointer">
            <input type="radio" name="send_mode" value="selected" class="text-primary-600">
            <span class="text-sm text-slate-700">
              {{if eq .Lang "fr"}}Envoyer uniquement aux sélectionnés ci-dessous{{else}}Send only to selected below{{end}}
            </span>
          </label>

          <div class="pt-3">
            <button type="submit" class="inline-flex items-center space-x-2 px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
              </svg>
              <span>{{if eq .Lang "fr"}}Envoyer les relances{{else}}Send Reminders{{end}}</span>
            </button>
          </div>
        </div>
      </div>
    </form>
    {{else}}
    <div class="text-center py-4 text-slate-500">
      <p>{{if eq .Lang "fr"}}✓ Tous les lecteurs attendus ont été contactés ou ont confirmé la lecture{{else}}✓ All expected readers have been contacted or have confirmed{{end}}</p>
    </div>
    {{end}}
  </div>
  {{end}}

  <!-- Unexpected Signatures Table -->
  {{if .UnexpectedSignatures}}
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-slate-900">
        {{if eq .Lang "fr"}}⚠ Confirmations de lecture non attendues{{else}}⚠ Unexpected Confirmations{{end}}
      </h2>
      <span class="text-sm font-medium text-slate-600">
        {{len .UnexpectedSignatures}}
      </span>
    </div>
    <p class="text-sm text-slate-600 mb-4">
      {{if eq .Lang "fr"}}Ces utilisateurs ont confirmé la lecture mais n'étaient pas dans la liste des lecteurs attendus.{{else}}These users confirmed reading but were not in the expected readers list.{{end}}
    </p>
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Lecteur{{else}}Reader{{end}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{if eq .Lang "fr"}}Confirmé le{{else}}Confirmed At{{end}}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-slate-200">
          {{range .UnexpectedSignatures}}
          <tr class="hover:bg-slate-50">
            <td class="px-6 py-4">
              <div class="flex items-center space-x-2">
                <div>
                  <span class="text-green-600">✓</span>
                </div>
                <div>
                  <div class="text-sm font-medium text-slate-900">
                    {{if .UserName}}
                      {{.UserName}} &lt;{{.UserEmail}}&gt;
                    {{else}}
                      {{.UserEmail}}
                    {{end}}
                  </div>
                </div>
              </div>
            </td>
            <td class="px-6 py-4 text-sm text-slate-500">
              <div>{{.SignedAtUTC.Format "02/01 15:04"}}</div>
            </td>
          </tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
  {{end}}

  <!-- Chain Integrity Section -->
  {{if .Signatures}}
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
          <strong>{{if eq .Lang "fr"}}Intégrité de la chaîne validée{{else}}Chain integrity valid{{end}}</strong> - {{.ChainIntegrity.ValidSigs}}/{{.ChainIntegrity.TotalSigs}} {{if eq .Lang "fr"}}confirmations{{else}}confirmations{{end}}
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
  {{end}}
</div>

<!-- Add Signers Modal -->
<div id="addSignersModal" class="hidden fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
    <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
      <div class="p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-xl font-semibold text-slate-900">
            {{if eq .Lang "fr"}}Ajouter des lecteurs attendus{{else}}Add Expected Readers{{end}}
          </h3>
          <button onclick="closeAddSignersModal()" class="text-slate-400 hover:text-slate-600">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <form method="POST" action="/admin/docs/{{.DocID}}/expected">
          <div class="mb-4">
            <label for="modalEmails" class="block text-sm font-medium text-slate-700 mb-2">
              {{if eq .Lang "fr"}}Lecteurs attendus (un par ligne){{else}}Expected Readers (one per line){{end}}
            </label>
            <textarea
              name="emails"
              id="modalEmails"
              rows="8"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="Marie Dupont <marie.dupont@example.com>&#10;jean.martin@example.com&#10;Sophie Bernard <sophie@example.com>"
            ></textarea>
            <p class="mt-2 text-xs text-slate-500">
              {{if eq .Lang "fr"}}Formats acceptés : "Nom Prénom &lt;email@example.com&gt;" ou "email@example.com"{{else}}Accepted formats: "First Last &lt;email@example.com&gt;" or "email@example.com"{{end}}
            </p>
          </div>

          <div class="flex justify-end space-x-3">
            <button type="button" onclick="closeAddSignersModal()" class="px-4 py-2 border border-slate-300 text-slate-700 font-medium rounded-lg hover:bg-slate-50 transition-colors">
              {{if eq .Lang "fr"}}Annuler{{else}}Cancel{{end}}
            </button>
            <button type="submit" class="px-4 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors">
              {{if eq .Lang "fr"}}Ajouter{{else}}Add{{end}}
            </button>
          </div>
        </form>
      </div>
    </div>
</div>

<!-- Edit Document Metadata Modal -->
<div id="editMetadataModal" class="hidden fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
  <div class="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
    <div class="p-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-xl font-semibold text-slate-900">
          {{if eq .Lang "fr"}}Modifier les métadonnées du document{{else}}Edit Document Metadata{{end}}
        </h3>
        <button onclick="closeEditMetadataModal()" class="text-slate-400 hover:text-slate-600">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <form method="POST" action="/admin/docs/{{.DocID}}/metadata" id="metadataForm">
        <div class="space-y-4">
          <div>
            <label for="metaTitle" class="block text-sm font-medium text-slate-700 mb-1">
              {{if eq .Lang "fr"}}Titre{{else}}Title{{end}}
            </label>
            <input
              type="text"
              name="title"
              id="metaTitle"
              value="{{if .Document}}{{.Document.Title}}{{end}}"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="{{if eq .Lang "fr"}}Politique de confidentialité 2025{{else}}Privacy Policy 2025{{end}}"
            >
          </div>

          <div>
            <label for="metaURL" class="block text-sm font-medium text-slate-700 mb-1">
              {{if eq .Lang "fr"}}URL / Emplacement{{else}}URL / Location{{end}}
            </label>
            <input
              type="text"
              name="url"
              id="metaURL"
              value="{{if .Document}}{{.Document.URL}}{{end}}"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="https://example.com/documents/policy.pdf"
            >
          </div>

          <div>
            <label for="metaChecksum" class="block text-sm font-medium text-slate-700 mb-1">
              {{if eq .Lang "fr"}}Empreinte (Checksum){{else}}Checksum{{end}}
            </label>
            <input
              type="text"
              name="checksum"
              id="metaChecksum"
              value="{{if .Document}}{{.Document.Checksum}}{{end}}"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono text-sm"
              placeholder="e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
            >
          </div>

          <div>
            <label for="metaAlgorithm" class="block text-sm font-medium text-slate-700 mb-1">
              {{if eq .Lang "fr"}}Algorithme{{else}}Algorithm{{end}}
            </label>
            <select
              name="checksum_algorithm"
              id="metaAlgorithm"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="SHA-256" {{if .Document}}{{if eq .Document.ChecksumAlgorithm "SHA-256"}}selected{{end}}{{else}}selected{{end}}>SHA-256</option>
              <option value="SHA-512" {{if .Document}}{{if eq .Document.ChecksumAlgorithm "SHA-512"}}selected{{end}}{{end}}>SHA-512</option>
              <option value="MD5" {{if .Document}}{{if eq .Document.ChecksumAlgorithm "MD5"}}selected{{end}}{{end}}>MD5</option>
            </select>
          </div>

          <div>
            <label for="metaDescription" class="block text-sm font-medium text-slate-700 mb-1">
              {{if eq .Lang "fr"}}Description{{else}}Description{{end}}
            </label>
            <textarea
              name="description"
              id="metaDescription"
              rows="4"
              class="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder="{{if eq .Lang "fr"}}Description optionnelle du document...{{else}}Optional document description...{{end}}"
            >{{if .Document}}{{.Document.Description}}{{end}}</textarea>
          </div>
        </div>

        <div class="flex justify-end space-x-3 mt-6">
          <button type="button" onclick="closeEditMetadataModal()" class="px-4 py-2 border border-slate-300 text-slate-700 font-medium rounded-lg hover:bg-slate-50 transition-colors">
            {{if eq .Lang "fr"}}Annuler{{else}}Cancel{{end}}
          </button>
          <button type="submit" class="px-4 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors">
            {{if eq .Lang "fr"}}Enregistrer{{else}}Save{{end}}
          </button>
        </div>
      </form>
    </div>
  </div>
</div>

<!-- Confirmation Modal -->
<div id="confirmModal" class="hidden fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
  <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
    <div class="p-6">
      <div class="flex items-center justify-center w-12 h-12 mx-auto mb-4 bg-orange-100 rounded-full">
        <svg class="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
      </div>

      <h3 class="text-lg font-semibold text-slate-900 text-center mb-2" id="confirmModalTitle">
        {{if eq .Lang "fr"}}Confirmation{{else}}Confirmation{{end}}
      </h3>

      <p class="text-sm text-slate-600 text-center mb-6" id="confirmModalMessage">
        <!-- Message will be set dynamically -->
      </p>

      <div class="flex justify-end space-x-3">
        <button type="button" onclick="closeConfirmModal()" class="px-4 py-2 border border-slate-300 text-slate-700 font-medium rounded-lg hover:bg-slate-50 transition-colors">
          {{if eq .Lang "fr"}}Annuler{{else}}Cancel{{end}}
        </button>
        <button type="button" id="confirmModalConfirm" onclick="confirmModalAction()" class="px-4 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors">
          {{if eq .Lang "fr"}}Confirmer{{else}}Confirm{{end}}
        </button>
      </div>
    </div>
  </div>
</div>

<!-- Delete Confirmation Modal -->
<div id="deleteModal" class="hidden fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
  <div class="bg-white rounded-lg shadow-xl max-w-md w-full">
    <div class="p-6">
      <div class="flex items-center justify-center w-12 h-12 mx-auto mb-4 bg-red-100 rounded-full">
        <svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
        </svg>
      </div>

      <h3 class="text-lg font-semibold text-slate-900 text-center mb-2">
        {{if eq .Lang "fr"}}Supprimer le lecteur attendu ?{{else}}Remove expected reader?{{end}}
      </h3>

      <p class="text-sm text-slate-600 text-center mb-6" id="deleteModalMessage">
        {{if eq .Lang "fr"}}Cette action est irréversible.{{else}}This action is irreversible.{{end}}
      </p>

      <div class="flex justify-end space-x-3">
        <button type="button" onclick="closeDeleteModal()" class="px-4 py-2 border border-slate-300 text-slate-700 font-medium rounded-lg hover:bg-slate-50 transition-colors">
          {{if eq .Lang "fr"}}Annuler{{else}}Cancel{{end}}
        </button>
        <button type="button" id="deleteModalConfirm" onclick="confirmDeleteAction()" class="px-4 py-2 bg-red-600 text-white font-medium rounded-lg hover:bg-red-700 transition-colors">
          {{if eq .Lang "fr"}}Supprimer{{else}}Delete{{end}}
        </button>
      </div>
    </div>
  </div>
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

function openAddSignersModal() {
  document.getElementById('addSignersModal').classList.remove('hidden');
  document.getElementById('modalEmails').focus();
}

function closeAddSignersModal() {
  document.getElementById('addSignersModal').classList.add('hidden');
  document.getElementById('modalEmails').value = '';
}

// Close modal on Escape key
document.addEventListener('keydown', function(e) {
  if (e.key === 'Escape') {
    closeAddSignersModal();
  }
});

// Close modal on backdrop click
document.getElementById('addSignersModal').addEventListener('click', function(e) {
  if (e.target === this) {
    closeAddSignersModal();
  }
});

// Toggle all checkboxes
function toggleSelectAll() {
  const selectAll = document.getElementById('selectAll');
  const checkboxes = document.querySelectorAll('.email-checkbox');
  checkboxes.forEach(cb => cb.checked = selectAll.checked);
}

// Modal management
let currentConfirmCallback = null;
let reminderForm = null;

function showConfirmModal(message, callback) {
  document.getElementById('confirmModalMessage').textContent = message;
  document.getElementById('confirmModal').classList.remove('hidden');
  currentConfirmCallback = callback;
}

function closeConfirmModal() {
  document.getElementById('confirmModal').classList.add('hidden');
  currentConfirmCallback = null;
}

function confirmModalAction() {
  if (currentConfirmCallback) {
    currentConfirmCallback();
  }
  closeConfirmModal();
}

// Delete modal
let currentDeleteForm = null;

function showDeleteModal(form) {
  document.getElementById('deleteModal').classList.remove('hidden');
  currentDeleteForm = form;
}

function closeDeleteModal() {
  document.getElementById('deleteModal').classList.add('hidden');
  currentDeleteForm = null;
}

function confirmDeleteAction() {
  if (currentDeleteForm) {
    currentDeleteForm.submit();
  }
  closeDeleteModal();
}

// Confirm before sending reminders
function confirmSendReminders() {
  const sendMode = document.querySelector('input[name="send_mode"]:checked');
  if (!sendMode) {
    showConfirmModal('{{if eq .Lang "fr"}}Veuillez sélectionner un mode d\'envoi{{else}}Please select a send mode{{end}}', null);
    return false;
  }

  const selectedCount = document.querySelectorAll('.email-checkbox:checked').length;
  let message;

  if (sendMode.value === 'all') {
    const pendingCount = {{.ReminderStats.PendingCount}};
    message = '{{if eq .Lang "fr"}}Confirmer l\'envoi de relances à{{else}}Confirm sending reminders to{{end}} ' + pendingCount + ' {{if eq .Lang "fr"}}lecteur(s) en attente ?{{else}}pending reader(s)?{{end}}';
  } else {
    if (selectedCount === 0) {
      showConfirmModal('{{if eq .Lang "fr"}}Veuillez sélectionner au moins un lecteur.{{else}}Please select at least one reader.{{end}}', null);
      return false;
    }
    message = '{{if eq .Lang "fr"}}Confirmer l\'envoi de relances à{{else}}Confirm sending reminders to{{end}} ' + selectedCount + ' {{if eq .Lang "fr"}}lecteur(s) sélectionné(s) ?{{else}}selected reader(s)?{{end}}';
  }

  // Store the form and show confirmation
  const form = event.target;
  reminderForm = form;
  showConfirmModal(message, function() {
    reminderForm.submit();
  });
  return false;
}

// Document metadata modal functions
function openEditMetadataModal() {
  document.getElementById('editMetadataModal').classList.remove('hidden');
}

function closeEditMetadataModal() {
  document.getElementById('editMetadataModal').classList.add('hidden');
}

// Copy checksum
function copyChecksum() {
  const checksumInput = document.getElementById('docChecksum');
  const feedback = document.getElementById('checksumCopyFeedback');

  checksumInput.select();
  checksumInput.setSelectionRange(0, 99999);

  navigator.clipboard.writeText(checksumInput.value).then(() => {
    feedback.classList.remove('hidden');
    setTimeout(() => {
      feedback.classList.add('hidden');
    }, 2000);
  });
}
</script>
{{end}}
