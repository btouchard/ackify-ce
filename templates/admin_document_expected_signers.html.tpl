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
              <form method="POST" action="/admin/docs/{{$.DocID}}/expected/remove" class="inline">
                <input type="hidden" name="email" value="{{.Email}}">
                <button type="submit" class="text-red-600 hover:text-red-900 font-medium" onclick="return confirm('{{if eq $.Lang "fr"}}Supprimer ce lecteur attendu ?{{else}}Remove this expected reader?{{end}}')">
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
          <!-- Document URL Input -->
          <div>
            <label for="docURL" class="block text-sm font-medium text-slate-700 mb-2">
              {{if eq .Lang "fr"}}URL ou emplacement du document{{else}}Document URL or location{{end}}
            </label>
            <input
              type="text"
              name="doc_url"
              id="docURL"
              placeholder="{{if eq .Lang "fr"}}https://example.com/doc.pdf ou /chemin/vers/document{{else}}https://example.com/doc.pdf or /path/to/document{{end}}"
              class="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg bg-white focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
            <p class="mt-1 text-xs text-slate-500">
              {{if eq .Lang "fr"}}Indiquez où se trouve le document à lire (URL ou chemin réseau){{else}}Specify where the document to read is located (URL or network path){{end}}
            </p>
          </div>

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

// Confirm before sending reminders
function confirmSendReminders() {
  const sendMode = document.querySelector('input[name="send_mode"]:checked');
  if (!sendMode) {
    alert('{{if eq .Lang "fr"}}Veuillez sélectionner un mode d\'envoi{{else}}Please select a send mode{{end}}');
    return false;
  }

  const selectedCount = document.querySelectorAll('.email-checkbox:checked').length;
  let message;

  if (sendMode.value === 'all') {
    const pendingCount = {{.ReminderStats.PendingCount}};
    message = '{{if eq .Lang "fr"}}Confirmer l\'envoi de relances à{{else}}Confirm sending reminders to{{end}} ' + pendingCount + ' {{if eq .Lang "fr"}}lecteur(s) en attente ?{{else}}pending reader(s)?{{end}}';
  } else {
    if (selectedCount === 0) {
      alert('{{if eq .Lang "fr"}}Veuillez sélectionner au moins un lecteur.{{else}}Please select at least one reader.{{end}}');
      return false;
    }
    message = '{{if eq .Lang "fr"}}Confirmer l\'envoi de relances à{{else}}Confirm sending reminders to{{end}} ' + selectedCount + ' {{if eq .Lang "fr"}}lecteur(s) sélectionné(s) ?{{else}}selected reader(s)?{{end}}';
  }

  return confirm(message);
}
</script>
{{end}}
