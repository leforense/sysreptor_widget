# CLAUDE.md — reptor_widget

## Objetivo do projeto

Dois binários Go independentes que integram o SysReptor ao desktop Kali Linux (XFCE), voltados para pentesters que trabalham com múltiplos projetos simultâneos:

- **`reptor_widget`** — plugin para o xfce4-genmon-plugin. Exibe o nome do projeto ativo na barra do painel em verde. Clique abre seletor de projeto via zenity. Atualiza o painel instantaneamente ao trocar de projeto.
- **`reptor-upload`** — hook post-capture do ksnip. Ao salvar um print, busca os findings do projeto ativo via API, exibe lista via zenity, faz upload multipart da imagem e atualiza (PATCH) o campo Markdown do finding selecionado. Memoriza o último finding usado por projeto.

---

## Arquitetura e decisões técnicas

### Dois binários separados — não um único com subcomandos
Decisão intencional para facilitar adoção pela comunidade. Cada binário tem responsabilidade única e pode ser compilado, copiado e testado de forma independente. Um único binário com flags (`reptor_widget upload`, `reptor_widget panel`) seria mais elegante tecnicamente mas adicionaria fricção para usuários que só precisam de um dos dois.

### Sem dependências externas
`go.mod` declara apenas stdlib. Nenhum `go get` necessário. Isso é crítico para o público-alvo (pentesters em Kali) que pode querer compilar em ambientes isolados ou sem acesso à internet.

### Código duplicado entre os dois arquivos — intencional
`WidgetConfig`, `parseField`, `loadWidgetConfig` e `loadConfig` estão duplicados em `main.go` e `ksnip-reptor.go`. A alternativa seria um package compartilhado (ex: `internal/config`), mas isso exigiria que o usuário entendesse a estrutura de módulos Go e usasse `go build ./...` em vez de `go build -o bin arquivo.go`. A duplicação de ~60 linhas foi a troca aceita para manter o build simples.

### Regex para parsing de config — sem biblioteca YAML
Ambos os arquivos de config (sysreptor `config.yaml` e widget `config.conf`) são lidos com `regexp.MustCompile` linha a linha. Isso elimina dependência de um parser YAML e é suficiente para o formato `key: value` simples usado. O `regexp.QuoteMeta(key)` garante segurança contra chaves com caracteres especiais.

### xfce4-panel refresh dinâmico
`forcePanelRefresh()` em `main.go` localiza o arquivo `genmon-*.rc` que referencia o executável pelo nome (`os.Args[0]`), extrai o plugin ID e dispara `xfce4-panel --plugin-event=genmon-N:refresh:bool:true`. Isso permite atualização instantânea sem esperar o timer de 30s do GenMon. Funciona mesmo que o usuário renomeie o binário.

### i18n sem biblioteca
Struct `T` com todos os campos de string + map `translations["pt-br"]` / `translations["en-us"]`. `getT(lang)` faz fallback para `en-us` se o idioma não for reconhecido. Format strings com `%v`, `%d`, `%s` ficam dentro da struct e são passados para `fmt.Errorf` / `fmt.Sprintf` nos call sites.

### Memória do último finding
Salvo em `~/.cache/reptor-widget/last_finding.json` como JSON `{"project_id":"...","finding_id":"..."}`. Verificação de projeto no load: se `project_id` não bater com o projeto atual, o cache é ignorado (evita crash ao trocar de projeto). Só é gravado após upload **bem-sucedido**.

### Fallback de autorização
`loadConfig(wc)` tenta primeiro `~/.sysreptor/config.yaml` (sysreptor CLI). Se `server` ou `token` estiverem ausentes, usa os valores de `WidgetConfig` (lidos de `~/.config/reptor-widget/config.conf`). Isso permite uso sem o sysreptor CLI descomentando linhas no arquivo de configuração.

---

## Estrutura de arquivos

```
reptor_widget/
├── main.go                          # binário: reptor_widget (panel widget)
├── go.mod                           # module reptor_widget, go 1.24.13
├── config.conf.example              # template de configuração para o usuário copiar
├── .gitignore                       # exclui binários compilados e *.png
├── README.md                        # documentação bilíngue com screenshots
├── CLAUDE.md                        # este arquivo
└── ksnip_addon_reptor/
    ├── ksnip-reptor.go              # binário: reptor-upload (ksnip hook)
    └── readme.txt                   # instrução de build legada (obsoleta, mantida)
```

### Arquivos de runtime (fora do repo)

| Arquivo | Propósito |
|---|---|
| `~/.sysreptor/config.yaml` | Config primária: `server`, `token`, `project_id` (sysreptor CLI) |
| `~/.config/reptor-widget/config.conf` | Config do widget: idioma, template_uuid, api_upload_field, server/token opcionais |
| `~/.cache/reptor-widget/last_finding.json` | Cache do último finding usado por projeto |

---

## Dependências e versões

- **Go:** 1.24.13 (declarado no go.mod; compatível com 1.21+ na prática)
- **zenity:** GUI de diálogos — já incluso no Kali Linux, sem instalação extra
- **xfce4-genmon-plugin:** necessário para o panel widget
- **ksnip:** necessário para o upload addon (hook post-capture)
- **xfce4-panel:** presente em qualquer instalação XFCE

Sem dependências Go externas. `go.sum` não existe.

---

## Padrões de código adotados

- Sem comentários exceto onde o comportamento é não-óbvio (ex: `forcePanelRefresh`, o match por `execName`)
- Funções de API HTTP seguem o padrão: criar request → setar header Authorization Bearer → executar com timeout → verificar status code → unmarshal JSON
- Erros de API sempre incluem status code e body na mensagem (para debug via zenity `--error`)
- `io.Copy` sem checagem de erro no upload (falha subsequente capturada no `writer.Close` ou no status code da resposta)
- Timeouts explícitos em todos os HTTP clients: 5s (widget display), 10s (listagens), 20s (upload)
- Campos de configuração com zero value são inválidos — `loadWidgetConfig` define defaults no struct literal

### Convenção de i18n
Adicionar novo idioma: inserir nova entrada no map `translations` em **ambos** os arquivos (`main.go` e `ksnip-reptor.go`). Campos de formato (`%v`, `%d`, `%s`) devem ter os mesmos placeholders e na mesma ordem que as chamadas nos call sites.

---

## O que já foi implementado

- [x] Panel widget funcional com refresh imediato ao trocar projeto
- [x] Upload de imagem via multipart form-data para a API do SysReptor
- [x] Seleção de finding existente ou criação via template
- [x] PATCH do campo Markdown com append (não sobrescreve conteúdo existente)
- [x] Memória do último finding por projeto (`~/.cache/reptor-widget/last_finding.json`)
- [x] Último finding aparece no topo da lista com marcador `★`
- [x] NEW_FINDING sempre no final da lista
- [x] i18n: `pt-br` e `en-us`
- [x] Crédito do desenvolvedor nos diálogos zenity
- [x] Arquivo de configuração do widget (`~/.config/reptor-widget/config.conf`)
- [x] `template_uuid` configurável (era hardcodado)
- [x] `api_upload_field` configurável (era hardcodado como `"api_upload"`)
- [x] Fallback de autorização: sysreptor CLI config → widget config
- [x] Validação de `template_uuid` antes de criar novo finding (erro claro se não configurado)
- [x] `.gitignore` excluindo binários e imagens de desenvolvimento

---

## O que ainda falta / limitações conhecidas

### Bug conhecido
Se `project_id` em `~/.sysreptor/config.yaml` estiver em branco ou apontar para um projeto inexistente/deletado, o widget exibe "Reptor Not Connected" sem distinção de "projeto inválido" vs "sem conexão". Workaround: editar o arquivo ou rodar `reptor project`.

### Auto-refresh ao trocar de projeto (parcialmente resolvido)
O `forcePanelRefresh()` funciona para atualização imediata após troca via clique. Mas se outra ferramenta alterar o `project_id` externamente (ex: `reptor project` no terminal), o widget só vai refletir na próxima chamada do GenMon timer (30s). Não há mecanismo de watch de arquivo.

### Sem instalador
O usuário precisa compilar e copiar os binários manualmente. Um `Makefile` ou script `install.sh` facilitaria o processo para a comunidade.

### ksnip_addon_reptor/readme.txt
Arquivo de instrução de build legado (3 linhas). Não conflita com nada, mas é redundante dado o README.md completo. Pode ser removido em algum momento.

### Sem testes automatizados
Não há nenhum arquivo `_test.go`. As funções de HTTP são difíceis de testar sem mock do servidor SysReptor. Se quiser adicionar testes, o caminho mais prático seria testes de integração contra uma instância local via Docker.

### Paginação da API não implementada
`getAllProjects` e `getFindings` consomem apenas a primeira página retornada pela API. Se o usuário tiver muitos projetos ou findings, os itens além do limite de página da API não aparecerão na lista.

---

## Contexto relevante da documentação

### SysReptor API
- Base URL: `{server}/api/v1/`
- Auth: header `Authorization: Bearer {token}`
- Endpoints usados:
  - `GET /pentestprojects/` → lista projetos (campo `results[]`)
  - `GET /pentestprojects/{id}` → detalhes do projeto (campo `name`)
  - `GET /pentestprojects/{id}/findings/` → lista findings (array direto, sem paginação wrapper)
  - `POST /pentestprojects/{id}/findings/fromtemplate/` → cria finding via template; body `{"template": "{uuid}"}` → retorna finding com `id`
  - `GET /pentestprojects/{id}/findings/{id}/` → detalhes do finding; campo `data` é `map[string]interface{}`
  - `PATCH /pentestprojects/{id}/findings/{id}/` → atualiza finding; body `{"data": {...}}`
  - `POST /pentestprojects/{id}/upload/` → upload multipart; campos `file` (binário) e `name` (string); retorna `{"name": "filename.png"}`
- Imagens referenciadas no Markdown como: `![alt](/images/name/{filename})`

### xfce4-genmon-plugin output format
O widget imprime para stdout no formato XML do GenMon:
```xml
<txt><span weight='Bold' fgcolor='Green'>Nome do Projeto</span></txt><txtclick>reptor_widget --menu</txtclick>
```
- `<txt>` → texto exibido no painel (suporta Pango markup)
- `<txtclick>` → comando executado ao clicar no widget

### ksnip post-capture script
O ksnip passa o caminho do arquivo salvo como `%1` (primeiro argumento do script). O arquivo já existe em disco quando o script é chamado.
