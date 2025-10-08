{{define "admin_dashboard"}}
<div class="space-y-6">
  <!-- Create Document Section -->
  <div class="bg-gradient-to-r from-primary-50 to-blue-50 rounded-lg shadow-sm border border-primary-200 p-6">
    <h2 class="text-lg font-semibold text-slate-900 mb-4">
      {{if eq .Lang "fr"}}Créer un nouveau document{{else}}Create New Document{{end}}
    </h2>
    <form id="createDocForm">
      <div class="flex flex-col sm:flex-row sm:items-start gap-4">
        <div class="flex-1">
          <label for="newDocId" class="block text-sm font-medium text-slate-700 mb-2">
            {{if eq .Lang "fr"}}ID du document{{else}}Document ID{{end}}
          </label>
          <input
            type="text"
            id="newDocId"
            name="doc_id"
            required
            pattern="[a-zA-Z0-9\-_]+"
            placeholder="{{if eq .Lang "fr"}}ex: politique-securite-2025{{else}}e.g. security-policy-2025{{end}}"
            class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
          <p class="mt-1 text-xs text-slate-500">
            {{if eq .Lang "fr"}}Lettres, chiffres, tirets et underscores uniquement{{else}}Letters, numbers, hyphens and underscores only{{end}}
          </p>
        </div>
        <div class="sm:pt-7">
          <button
            type="submit"
            class="w-full sm:w-auto px-6 py-2 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors inline-flex items-center justify-center space-x-2 whitespace-nowrap"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
            </svg>
            <span>{{if eq .Lang "fr"}}Créer{{else}}Create{{end}}</span>
          </button>
        </div>
      </div>
    </form>
  </div>

  <script>
    document.getElementById('createDocForm').addEventListener('submit', function(e) {
      e.preventDefault();
      const docId = document.getElementById('newDocId').value.trim();
      if (docId) {
        window.location.href = '/admin/docs/' + encodeURIComponent(docId);
      }
    });
  </script>

  <!-- Documents List -->
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-slate-900">{{index .T "admin.title"}}</h1>
      <p class="text-slate-600 mt-1">{{index .T "admin.subtitle"}}</p>
    </div>

    {{if .Documents}}
    <div class="overflow-hidden">
      <table class="min-w-full divide-y divide-slate-200">
        <thead class="bg-slate-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{index .T "admin.doc_id"}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{index .T "admin.signatures_count"}}
            </th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase tracking-wider">
              {{index .T "admin.actions"}}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-slate-200">
          {{range .Documents}}
          <tr class="hover:bg-slate-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-slate-900">{{.DocID}}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm text-slate-900">
                {{if gt .ExpectedCount 0}}
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                    {{.Count}} {{if eq .Count 1}}{{index $.T "admin.signature_singular"}}{{else}}{{index $.T "admin.signature_plural"}}{{end}}
                    {{if gt .UnexpectedCount 0}} (+{{.UnexpectedCount}}){{end}}
                    sur {{.ExpectedCount}}
                  </span>
                {{else}}
                  <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                    {{.Count}} {{if ne .Count 1}}{{index $.T "admin.signature_plural"}}{{else}}{{index $.T "admin.signature_singular"}}{{end}}
                  </span>
                {{end}}
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
              <a href="/admin/docs/{{.DocID}}" class="text-primary-600 hover:text-primary-900">
                {{index $.T "admin.view_details"}}
              </a>
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
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
        </svg>
      </div>
      <h3 class="text-lg font-medium text-slate-900 mb-2">{{index .T "admin.no_docs_title"}}</h3>
      <p class="text-slate-600">{{index .T "admin.no_docs_desc"}}</p>
    </div>
    {{end}}
  </div>
</div>
{{end}}