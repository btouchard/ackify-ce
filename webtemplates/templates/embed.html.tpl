{{define "embed"}}<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Signataires - Document {{.DocID}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        html, body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #ffffff;
            color: #334155;
            line-height: 1.4;
            padding: 0;
            margin: 0;
            height: 100%;
            overflow-x: hidden;
        }
        
        .embed-container {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
            border: 1px solid #e2e8f0;
            overflow: hidden;
            width: 100%;
            height: 100%;
            min-width: 280px;
            max-width: 100%;
            display: flex;
            flex-direction: column;
        }
        
        .header {
            background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
            color: white;
            padding: 10px 16px;
            border-bottom: 1px solid #e2e8f0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .header h3 {
            font-size: 16px;
            font-weight: 600;
            margin: 0;
            display: flex;
            align-items: center;
            gap: 6px;
            flex: 1;
        }
        
        .header .parent-domain {
            font-size: 11px;
            opacity: 0.8;
            text-align: right;
            flex-shrink: 0;
        }
        
        .header .doc-id {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            background: rgba(255, 255, 255, 0.2);
            padding: 3px 6px;
            border-radius: 4px;
            font-size: 12px;
            word-break: break-all;
            max-width: 120px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .stats {
            background: #f8fafc;
            padding: 10px 16px;
            border-bottom: 1px solid #e2e8f0;
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            font-size: 13px;
            gap: 8px;
            flex-wrap: wrap;
        }
        
        .stats .count {
            font-weight: 600;
            color: #059669;
        }
        
        .stats .last-signed {
            color: #6b7280;
            text-align: right;
            flex-shrink: 0;
        }
        
        .signatories {
            flex: 1;
            overflow-y: auto;
            min-height: 0;
        }
        
        .signatory {
            display: flex;
            align-items: center;
            padding: 10px 16px;
            border-bottom: 1px solid #f1f5f9;
        }
        
        .signatory:last-child {
            border-bottom: none;
        }
        
        .signatory:hover {
            background: #f8fafc;
        }
        
        .signatory-info {
            flex: 1;
            min-width: 0;
        }
        
        .signatory-email {
            font-weight: 500;
            color: #1e293b;
            font-size: 13px;
            word-break: break-word;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
        
        .signatory-date {
            color: #64748b;
            font-size: 11px;
            margin-top: 2px;
        }
        
        .signature-icon {
            width: 24px;
            height: 24px;
            background: #dcfce7;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 10px;
            flex-shrink: 0;
        }
        
        .signature-icon svg {
            width: 12px;
            height: 12px;
            color: #059669;
        }
        
        .empty-state {
            flex: 1;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            padding: 30px 16px;
            text-align: center;
            color: #64748b;
            min-height: 0;
        }
        
        .empty-state svg {
            width: 40px;
            height: 40px;
            color: #cbd5e1;
            margin-bottom: 10px;
        }
        
        .empty-state p {
            font-size: 14px;
            margin-bottom: 4px;
        }
        
        .footer {
            background: #f8fafc;
            padding: 10px 16px;
            text-align: center;
            border-top: 1px solid #e2e8f0;
        }
        
        .sign-button {
            background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
            color: white;
            text-decoration: none;
            font-size: 13px;
            font-weight: 600;
            display: inline-block;
            padding: 10px 20px;
            border-radius: 6px;
            border: none;
            cursor: pointer;
            transition: all 0.2s ease;
            box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
        }
        
        .sign-button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 8px rgba(59, 130, 246, 0.3);
            text-decoration: none;
            color: white;
        }
        
        .sign-button:active {
            transform: translateY(0px);
            box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
        }
        
        /* Scrollbar styling */
        .signatories::-webkit-scrollbar {
            width: 3px;
        }
        
        .signatories::-webkit-scrollbar-track {
            background: #f1f5f9;
        }
        
        .signatories::-webkit-scrollbar-thumb {
            background: #cbd5e1;
            border-radius: 2px;
        }
        
        .signatories::-webkit-scrollbar-thumb:hover {
            background: #94a3b8;
        }
        
        /* Responsive design for very narrow screens */
        @media (max-width: 320px) {
            .header {
                padding: 8px 12px;
                flex-direction: column;
                align-items: flex-start;
                gap: 4px;
            }
            
            .header .parent-domain {
                text-align: left;
            }
            
            .header h3 {
                font-size: 14px;
                gap: 4px;
            }
            
            .header .doc-id {
                font-size: 11px;
                max-width: 100px;
            }
            
            .stats {
                padding: 8px 12px;
                font-size: 12px;
                flex-direction: column;
                align-items: flex-start;
                gap: 4px;
            }
            
            .signatory {
                padding: 8px 12px;
            }
            
            .signatory-email {
                font-size: 12px;
            }
            
            .signatory-date {
                font-size: 10px;
            }
            
            .signature-icon {
                width: 20px;
                height: 20px;
                margin-right: 8px;
            }
            
            .signature-icon svg {
                width: 10px;
                height: 10px;
            }
            
            .footer {
                padding: 8px 12px;
            }
            
            .sign-button {
                font-size: 12px;
                padding: 8px 16px;
            }
            
            .empty-state {
                padding: 20px 12px;
            }
        }
        
        /* Google Drive sidebar specific optimizations */
        @media (max-width: 400px) {
            .embed-container {
                border-radius: 4px;
                box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
            }
            
            .header h3 {
                font-size: 15px;
            }
            
            .stats .last-signed {
                font-size: 11px;
                line-height: 1.3;
            }
            
            .signatories {
                flex: 1;
                min-height: 0;
            }
        }
        
        /* Compact mode for iframe embedding */
        .compact .signatories {
            flex: 1;
            min-height: 0;
        }
        
        .compact .empty-state {
            padding: 20px 16px;
            flex: 1;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
        }
        
        .compact .empty-state svg {
            width: 32px;
            height: 32px;
        }
    </style>
</head>
<body>
    <div class="embed-container">
        <div class="header">
            <h3>
                <svg width="20" height="20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                Signataires
                <span class="doc-id">{{.DocID}}</span>
            </h3>
            <div id="parent-domain" class="parent-domain"></div>
        </div>
        
        {{if gt .Count 0}}
            <div class="stats">
                <span class="count">{{.Count}} signature{{if gt .Count 1}}s{{end}}</span>
                {{if .LastSignedAt}}
                    <span class="last-signed">Dernière signature le {{.LastSignedAt}}</span>
                {{end}}
            </div>
            
            <div class="signatories">
                {{range .Signatures}}
                    <div class="signatory">
                        <div class="signature-icon">
                            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                            </svg>
                        </div>
                        <div class="signatory-info">
                            <div class="signatory-email">{{if .Name}}{{.Name}} • {{end}}{{.Email}}</div>
                            <div class="signatory-date">{{.SignedAt}}</div>
                        </div>
                    </div>
                {{end}}
            </div>
        {{else}}
            <div class="empty-state">
                <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
                <p><strong>Aucune signature</strong></p>
                <p>Ce document n'a pas encore été signé.</p>
            </div>
        {{end}}
        
        <div class="footer">
            <a href="{{$.SignURL}}" target="_blank" class="sign-button">
                Signer et confirmer la lecture de ce document
            </a>
        </div>
    </div>

    <script>
        // Variable globale pour stocker les infos du referrer détecté
        let detectedReferrer = null;
        
        // Détecter le domaine parent de l'iframe
        function detectParentDomain() {
            const parentDomainEl = document.getElementById('parent-domain');
            
            try {
                // Essayer d'accéder au domaine parent
                let parentHost = '';
                let parentOrigin = '';
                
                // Vérifier si on est dans un iframe
                if (window.parent !== window.self) {
                    try {
                        // Tenter d'accéder à l'URL du parent (peut échouer à cause de CORS)
                        parentHost = window.parent.location.hostname;
                        parentOrigin = window.parent.location.origin;
                    } catch (e) {
                        // Si bloqué par CORS, essayer avec document.referrer
                        if (document.referrer) {
                            try {
                                const referrerUrl = new URL(document.referrer);
                                parentHost = referrerUrl.hostname;
                                parentOrigin = referrerUrl.origin;
                            } catch (err) {
                                console.log('Impossible de parser le referrer:', err);
                            }
                        }
                    }
                    
                    // Afficher les informations si disponibles
                    if (parentHost) {
                        // Détecter le service basé sur le domaine
                        let serviceInfo = detectService(parentHost);
                        
                        if (serviceInfo) {
                            parentDomainEl.innerHTML = `${serviceInfo.icon} Intégré dans ${serviceInfo.name}`;
                            // Stocker les infos du referrer pour l'URL de signature
                            detectedReferrer = serviceInfo.referrer;
                        } else {
                            parentDomainEl.innerHTML = `🌐 Intégré dans ${parentHost}`;
                            // Utiliser le domaine nettoyé comme referrer
                            detectedReferrer = parentHost.replace(/[^a-z0-9]/g, '-');
                        }
                        
                        // Ajouter l'information comme attribut pour debugging
                        parentDomainEl.setAttribute('data-parent-domain', parentHost);
                        parentDomainEl.setAttribute('data-parent-origin', parentOrigin);
                        parentDomainEl.setAttribute('data-referrer', detectedReferrer);
                    } else {
                        parentDomainEl.innerHTML = '📱 Intégré (origine non détectable)';
                    }
                } else {
                    // Pas dans un iframe
                    parentDomainEl.innerHTML = '🌐 Vue directe';
                }
            } catch (e) {
                console.log('Erreur lors de la détection du domaine parent:', e);
                parentDomainEl.innerHTML = '🔒 Origine protégée';
            }
        }
        
        // Fonction pour détecter le service basé sur le hostname
        function detectService(hostname) {
            const host = hostname.toLowerCase();
            
            // Google services (including script.googleusercontent.com)
            if (host.includes('docs.google.com')) {
                return { name: 'Google Docs', icon: '📝', referrer: 'google-docs' };
            }
            if (host.includes('sheets.google.com')) {
                return { name: 'Google Sheets', icon: '📊', referrer: 'google-sheets' };
            }
            if (host.includes('slides.google.com')) {
                return { name: 'Google Slides', icon: '📊', referrer: 'google-slides' };
            }
            if (host.includes('drive.google.com')) {
                return { name: 'Google Drive', icon: '💾', referrer: 'google-drive' };
            }
            if (host.includes('script.googleusercontent.com') || host.includes('googleusercontent.com')) {
                return { name: 'Google', icon: '🔵', referrer: 'google' };
            }
            if (host.includes('google.com')) {
                return { name: 'Google', icon: '🔵', referrer: 'google' };
            }
            
            // Notion
            if (host.includes('notion.so') || host.includes('notion.com')) {
                return { name: 'Notion', icon: '📒', referrer: 'notion' };
            }
            
            // Confluence
            if (host.includes('confluence')) {
                return { name: 'Confluence', icon: '🌊', referrer: 'confluence' };
            }
            
            // Microsoft Office
            if (host.includes('office.com') || host.includes('sharepoint.com')) {
                return { name: 'Microsoft Office', icon: '🏢', referrer: 'microsoft' };
            }
            if (host.includes('live.com') || host.includes('outlook.com')) {
                return { name: 'Microsoft', icon: '🏢', referrer: 'microsoft' };
            }
            
            // GitHub
            if (host.includes('github.com')) {
                return { name: 'GitHub', icon: '🐙', referrer: 'github' };
            }
            
            // GitLab
            if (host.includes('gitlab.com')) {
                return { name: 'GitLab', icon: '🦊', referrer: 'gitlab' };
            }
            if (host.includes('gitlab')) {
                return { name: 'GitLab', icon: '🦊', referrer: 'gitlab' };
            }
            
            // Outline
            if (host.includes('outline')) {
                return { name: 'Outline', icon: '📖', referrer: 'outline' };
            }
            
            // Slack
            if (host.includes('slack.com')) {
                return { name: 'Slack', icon: '💬', referrer: 'slack' };
            }
            
            // Discord
            if (host.includes('discord.com')) {
                return { name: 'Discord', icon: '💬', referrer: 'discord' };
            }
            
            // Trello
            if (host.includes('trello.com')) {
                return { name: 'Trello', icon: '📋', referrer: 'trello' };
            }
            
            // Asana
            if (host.includes('asana.com')) {
                return { name: 'Asana', icon: '✅', referrer: 'asana' };
            }
            
            // Monday.com
            if (host.includes('monday.com')) {
                return { name: 'Monday.com', icon: '📅', referrer: 'monday' };
            }
            
            // Figma
            if (host.includes('figma.com')) {
                return { name: 'Figma', icon: '🎨', referrer: 'figma' };
            }
            
            // Miro
            if (host.includes('miro.com')) {
                return { name: 'Miro', icon: '🎨', referrer: 'miro' };
            }
            
            // Dropbox
            if (host.includes('dropbox.com')) {
                return { name: 'Dropbox', icon: '📦', referrer: 'dropbox' };
            }
            
            // Unknown service - use domain as referrer
            return { name: host, icon: '🌐', referrer: host.replace(/[^a-z0-9]/g, '-') };
        }
        
        // Fonction pour mettre à jour l'URL de signature avec le referrer
        function updateSignatureURL() {
            const signButton = document.querySelector('.sign-button');
            if (signButton && detectedReferrer) {
                const currentUrl = new URL(signButton.href);
                currentUrl.searchParams.set('referrer', detectedReferrer);
                signButton.href = currentUrl.toString();
            }
        }
        
        // Détecter le domaine parent au chargement de la page
        document.addEventListener('DOMContentLoaded', function() {
            detectParentDomain();
            // Petite pause pour s'assurer que detectedReferrer est défini
            setTimeout(updateSignatureURL, 150);
        });
        
        // Retry après un court délai au cas où les permissions changeraient
        setTimeout(function() {
            detectParentDomain();
            updateSignatureURL();
        }, 100);
    </script>
</body>
</html>{{end}}