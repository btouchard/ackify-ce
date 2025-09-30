{{define "admin_dashboard"}}
<div class="space-y-6">
  <div class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold text-slate-900">{{index .T "admin.title"}}</h1>
        <p class="text-slate-600 mt-1">{{index .T "admin.subtitle"}}</p>
      </div>
      <div class="flex items-center space-x-2">
        <div class="w-3 h-3 bg-green-500 rounded-full"></div>
        <span class="text-sm text-slate-600">{{index .T "admin.connected"}}</span>
      </div>
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
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary-100 text-primary-800">
                  {{.Count}} {{if eq $.Lang "fr"}}signature{{if ne .Count 1}}s{{end}}{{else}}signature{{if ne .Count 1}}s{{end}}{{end}}
                </span>
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