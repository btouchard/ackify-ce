// SPDX-License-Identifier: AGPL-3.0-or-later
function onOpen() {
  DocumentApp.getUi()
    .createMenu("Signatures")
    // .addItem("Confirmer la lecture de ce document", "openSignature")
    .addItem("Afficher la barre latérale", "showSignatures")
    .addToUi();
}

function getSidebarHtml() {
  var doc = DocumentApp.getActiveDocument();
  var docId = doc.getId();
  var url = "https://sign.neodtx.com/embed?doc=" + encodeURIComponent(docId);

  var response = UrlFetchApp.fetch(url, {muteHttpExceptions: true});
  return response.getContentText();
}

function openSignature() {
  var doc = DocumentApp.getActiveDocument();
  var docId = doc.getId();
  var url = "https://sign.neodtx.com/sign?doc=" + encodeURIComponent(docId);

  var html = '<style>html,body{height:100%;margin:0;padding:0}</style>' +
            '<iframe src="' + url + '" ' +
            'style="border:0;width:100%;height:100vh;" ' +
            'sandbox="allow-scripts allow-popups allow-same-origin allow-forms"></iframe>';

  var output = HtmlService.createHtmlOutput(html);
  DocumentApp.getUi().showModalDialog(output, 'Confirmer la lecture du document');
}

function showSignatures() {
  var doc = DocumentApp.getActiveDocument();
  var docId = doc.getId();
  var url = "https://sign.neodtx.com/embed?doc=" + encodeURIComponent(docId);

  var response = UrlFetchApp.fetch(url, {muteHttpExceptions: true});
  var html = response.getContentText();

  var modifiedHtml = html + `
  <script>
  document.addEventListener('DOMContentLoaded', function() {
    // Fonction pour rafraîchir le sidebar
    function refreshSidebar() {
      google.script.run.withSuccessHandler(function(newHtml){
        document.body.innerHTML = newHtml;
        // Réinjecte les listeners après rafraîchissement
        addLinkListeners();
      }).getSidebarHtml();
    }

    // Ajoute les listeners sur tous les liens
    function addLinkListeners() {
      document.querySelectorAll('a[href]').forEach(function(link){
        link.addEventListener('click', function(e){
          e.preventDefault();
          // Ajoute un listener focus sur window
          function onFocus() {
            window.removeEventListener('focus', onFocus); // on supprime après déclenchement
            refreshSidebar();
          }
          window.addEventListener('focus', onFocus);
          window.open(link.href, '_blank');
        });
      });
    }

    addLinkListeners(); // initial call
  });
  </script>
  `;

  var output = HtmlService.createHtmlOutput(modifiedHtml)
    .setTitle("Signatures")
    .setXFrameOptionsMode(HtmlService.XFrameOptionsMode.ALLOWALL)
    .setSandboxMode(HtmlService.SandboxMode.IFRAME);

  DocumentApp.getUi().showSidebar(output);
}

// function showSignatures() {
//   var doc = DocumentApp.getActiveDocument();
//   var docId = doc.getId();
//   var url = "https://sign.neodtx.com/embed?doc=" + encodeURIComponent(docId);

//   // On insère un iframe pointant sur ton embed
//   var html = '<style>html,body{height:100%;margin:0;padding:0}</style>' +
//              '<iframe src="' + url + '" ' +
//              'style="border:0;width:100%;height:100vh;" ' +
//              'sandbox="allow-scripts allow-popups allow-same-origin allow-forms"></iframe>';

//   var output = HtmlService.createHtmlOutput(html)
//               .setTitle("Signatures du document")
//               .setWidth(360); // largeur sidebar (modifiable)

//   DocumentApp.getUi().showSidebar(output);
// }
