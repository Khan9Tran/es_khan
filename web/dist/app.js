/**
 * ESKhan - Elasticsearch Query IDE Frontend Application
 * Pro / Premium Grade with Query Linter, Context-Aware Autocomplete,
 * Aggregations Visualizer, and Document CRUD.
 */

(function () {
  'use strict';

  // State Management
  const state = {
    activeCluster: null,
    clusters: [],
    indices: [],
    activeIdx: '',
    mappingCache: {},
    topTermsCache: {},
    dslSnippets: [],
    safeMode: true,
    linterWarnings: [],
    linterTimeout: null,
    tabs: [
      {
        id: 1,
        title: 'Query 1',
        index: '',
        content: '{\n  "size": 20,\n  "query": {\n    "match_all": {}\n  }\n}'
      }
    ],
    activeTabId: 1,
    lastResult: null,
    editorInstance: null,
    isMonacoLoaded: false,
    currentView: 'tree', // tree, table, chart, raw
    currentInspectDoc: null,
    theme: localStorage.getItem('eskhan_theme') || 'dark',
    currentLang: localStorage.getItem('eskhan_lang') || 'vi',
    jsonScope: 'hits' // 'hits' (focus on documents) or 'full' (complete response)
  };

  // DOM Elements
  const el = {
    clusterSelect: document.getElementById('clusterSelect'),
    clusterStatusBadge: document.getElementById('clusterStatusBadge'),
    clusterStatusText: document.getElementById('clusterStatusText'),
    btnManageClusters: document.getElementById('btnManageClusters'),
    btnSafeMode: document.getElementById('btnSafeMode'),
    safeModeLabel: document.getElementById('safeModeLabel'),
    btnToggleLang: document.getElementById('btnToggleLang'),
    langFlag: document.getElementById('langFlag'),
    langLabel: document.getElementById('langLabel'),
    btnToggleTheme: document.getElementById('btnToggleTheme'),
    themeIconSun: document.getElementById('themeIconSun'),
    themeIconMoon: document.getElementById('themeIconMoon'),
    editorIndexSelect: document.getElementById('editorIndexSelect'),
    btnInsertEndpoint: document.getElementById('btnInsertEndpoint'),
    btnScopeHits: document.getElementById('btnScopeHits'),
    btnScopeFull: document.getElementById('btnScopeFull'),
    linterBadge: document.getElementById('linterBadge'),
    linterCount: document.getElementById('linterCount'),
    btnFormat: document.getElementById('btnFormat'),
    btnSaveSnippetModal: document.getElementById('btnSaveSnippetModal'),
    btnRunQuery: document.getElementById('btnRunQuery'),
    sidebarTabs: document.querySelectorAll('.sidebar-tab'),
    sidebarContents: document.querySelectorAll('.sidebar-content'),
    indicesList: document.getElementById('indicesList'),
    indicesSearchInput: document.getElementById('indicesSearchInput'),
    btnRefreshIndices: document.getElementById('btnRefreshIndices'),
    snippetsList: document.getElementById('snippetsList'),
    snippetsSearchInput: document.getElementById('snippetsSearchInput'),
    historyList: document.getElementById('historyList'),
    btnClearHistory: document.getElementById('btnClearHistory'),
    editorTabsList: document.getElementById('editorTabsList'),
    btnNewTab: document.getElementById('btnNewTab'),
    editorContainer: document.getElementById('monacoEditorContainer'),
    editorSyntaxStatus: document.getElementById('editorSyntaxStatus'),
    paneResizer: document.getElementById('paneResizer'),
    editorPane: document.querySelector('.editor-pane'),
    viewTabs: document.querySelectorAll('.view-tab'),
    resultPanels: document.querySelectorAll('.result-view-panel'),
    tabViewChart: document.getElementById('tabViewChart'),
    metricStatus: document.getElementById('metricStatus'),
    metricTook: document.getElementById('metricTook'),
    metricLatency: document.getElementById('metricLatency'),
    metricHits: document.getElementById('metricHits'),
    metricShards: document.getElementById('metricShards'),
    btnCopyResult: document.getElementById('btnCopyResult'),
    btnExportCSV: document.getElementById('btnExportCSV'),
    jsonTreeContainer: document.getElementById('jsonTreeContainer'),
    tableContainer: document.getElementById('tableContainer'),
    aggChartContainer: document.getElementById('aggChartContainer'),
    rawJsonViewer: document.getElementById('rawJsonViewer'),

    // Document Inspector
    docInspectorModal: document.getElementById('docInspectorModal'),
    btnCloseDocInspector: document.getElementById('btnCloseDocInspector'),
    docInspectIndex: document.getElementById('docInspectIndex'),
    docInspectId: document.getElementById('docInspectId'),
    docInspectSafeNotice: document.getElementById('docInspectSafeNotice'),
    docInspectContent: document.getElementById('docInspectContent'),
    btnSaveDoc: document.getElementById('btnSaveDoc'),
    btnDeleteDoc: document.getElementById('btnDeleteDoc'),

    // Linter Modal
    linterModal: document.getElementById('linterModal'),
    btnCloseLinterModal: document.getElementById('btnCloseLinterModal'),
    linterWarningsList: document.getElementById('linterWarningsList'),

    // Cluster & Snippet Modals
    clusterModal: document.getElementById('clusterModal'),
    btnCloseClusterModal: document.getElementById('btnCloseClusterModal'),
    clusterProfilesList: document.getElementById('clusterProfilesList'),
    clusterForm: document.getElementById('clusterForm'),
    btnTestConn: document.getElementById('btnTestConn'),
    testConnResult: document.getElementById('testConnResult'),
    connAuthType: document.getElementById('connAuthType'),

    saveSnippetModal: document.getElementById('saveSnippetModal'),
    btnCloseSnippetModal: document.getElementById('btnCloseSnippetModal'),
    btnCancelSnippet: document.getElementById('btnCancelSnippet'),
    saveSnippetForm: document.getElementById('saveSnippetForm'),

    mappingModal: document.getElementById('mappingModal'),
    btnCloseMappingModal: document.getElementById('btnCloseMappingModal'),
    mappingModalTitle: document.getElementById('mappingModalTitle'),
    mappingFilterInput: document.getElementById('mappingFilterInput'),
    mappingFieldsBody: document.getElementById('mappingFieldsBody'),
    btnInsertTemplateQuery: document.getElementById('btnInsertTemplateQuery'),

    // Antigravity
    btnAntigravity: document.getElementById('btnAntigravity'),
    antigravityDot: document.getElementById('antigravityDot'),
    aiPromptBar: document.getElementById('aiPromptBar'),
    aiActiveIndexBadge: document.getElementById('aiActiveIndexBadge'),
    aiPromptInput: document.getElementById('aiPromptInput'),
    btnAIGenerate: document.getElementById('btnAIGenerate'),
    btnCloseAIPromptBar: document.getElementById('btnCloseAIPromptBar'),
    aiLoadingIndicator: document.getElementById('aiLoadingIndicator'),
    aiLoadingText: document.getElementById('aiLoadingText'),
    aiClarificationBox: document.getElementById('aiClarificationBox'),
    aiClarificationQuestion: document.getElementById('aiClarificationQuestion'),
    aiClarificationOptions: document.getElementById('aiClarificationOptions'),
    aiVerifiedBox: document.getElementById('aiVerifiedBox'),
    aiVerifiedText: document.getElementById('aiVerifiedText'),
    btnAIExplain: document.getElementById('btnAIExplain'),
    antigravityModal: document.getElementById('antigravityModal'),
    btnCloseAntigravityModal: document.getElementById('btnCloseAntigravityModal'),
    agModalStatusBadge: document.getElementById('agModalStatusBadge'),
    agModalPath: document.getElementById('agModalPath'),
    btnCheckAgyStatus: document.getElementById('btnCheckAgyStatus'),
    btnOpenAIPromptFromModal: document.getElementById('btnOpenAIPromptFromModal'),
    aiExplainModal: document.getElementById('aiExplainModal'),
    btnCloseExplainModal: document.getElementById('btnCloseExplainModal'),
    aiExplainContent: document.getElementById('aiExplainContent'),

    // Golang DSL Converter
    btnGolangModal: document.getElementById('btnGolangModal'),
    golangModal: document.getElementById('golangModal'),
    btnCloseGolangModal: document.getElementById('btnCloseGolangModal'),
    tabModeQueryToGo: document.getElementById('tabModeQueryToGo'),
    tabModeGoToQuery: document.getElementById('tabModeGoToQuery'),
    panelQueryToGo: document.getElementById('panelQueryToGo'),
    panelGoToQuery: document.getElementById('panelGoToQuery'),
    golangGeneratedOutput: document.getElementById('golangGeneratedOutput'),
    btnCopyGolangCode: document.getElementById('btnCopyGolangCode'),
    labelCopyGoCode: document.getElementById('labelCopyGoCode'),
    btnCopyGolangStructs: document.getElementById('btnCopyGolangStructs'),
    golangInputCode: document.getElementById('golangInputCode'),
    btnConvertGoToDSL: document.getElementById('btnConvertGoToDSL'),
    btnApplyDSLToEditor: document.getElementById('btnApplyDSLToEditor'),
    dslOutputWrapper: document.getElementById('dslOutputWrapper'),
    dslGeneratedOutput: document.getElementById('dslGeneratedOutput'),

    // v2: Mode Switcher & gRPC Studio
    btnModeES: document.getElementById('btnModeES'),
    btnModeGRPC: document.getElementById('btnModeGRPC'),
    clusterSelectorContainer: document.getElementById('clusterSelectorContainer'),
    esWorkspace: document.getElementById('esWorkspace'),
    grpcWorkspace: document.getElementById('grpcWorkspace'),
    grpcTargetInput: document.getElementById('grpcTargetInput'),
    grpcPlaintextCheck: document.getElementById('grpcPlaintextCheck'),
    grpcInsecureCheck: document.getElementById('grpcInsecureCheck'),
    btnGrpcReflect: document.getElementById('btnGrpcReflect'),
    btnOpenGrpcProtoModal: document.getElementById('btnOpenGrpcProtoModal'),
    grpcServiceSelect: document.getElementById('grpcServiceSelect'),
    grpcMethodSelect: document.getElementById('grpcMethodSelect'),
    btnGrpcInvoke: document.getElementById('btnGrpcInvoke'),
    btnGrpcFormatBody: document.getElementById('btnGrpcFormatBody'),
    btnGrpcResetMock: document.getElementById('btnGrpcResetMock'),
    grpcMonacoContainer: document.getElementById('grpcMonacoContainer'),
    grpcMetaCount: document.getElementById('grpcMetaCount'),
    grpcMetaTable: document.getElementById('grpcMetaTable'),
    grpcMetaRows: document.getElementById('grpcMetaRows'),
    btnAddMetaRow: document.getElementById('btnAddMetaRow'),
    grpcTimeoutInput: document.getElementById('grpcTimeoutInput'),
    grpcSplitResizer: document.getElementById('grpcSplitResizer'),
    grpcStatusBadge: document.getElementById('grpcStatusBadge'),
    grpcStatusDot: document.getElementById('grpcStatusDot'),
    grpcStatusText: document.getElementById('grpcStatusText'),
    grpcLatencyVal: document.getElementById('grpcLatencyVal'),
    grpcSizeVal: document.getElementById('grpcSizeVal'),
    btnCopyGrpcResponse: document.getElementById('btnCopyGrpcResponse'),
    grpcRespMonacoContainer: document.getElementById('grpcRespMonacoContainer'),
    grpcRespHeadersList: document.getElementById('grpcRespHeadersList'),
    grpcProtoModal: document.getElementById('grpcProtoModal'),
    btnCloseGrpcProtoModal: document.getElementById('btnCloseGrpcProtoModal'),
    btnCancelGrpcProto: document.getElementById('btnCancelGrpcProto'),
    grpcProtoInput: document.getElementById('grpcProtoInput'),
    btnParseProtoSubmit: document.getElementById('btnParseProtoSubmit')
  };

  // --- Bilingual Localization (VI / EN) ---
  const I18N = {
    vi: {
      lang_flag: "🇻🇳",
      lang_label: "VI",
      btn_lang_title: "Chuyển sang Tiếng Anh (Switch to English)",
      cluster_health: "Trạng thái Cụm",
      cluster_profiles: "Chọn Cấu hình Cụm",
      cluster_connecting: "Đang kết nối...",
      manage_clusters: "Quản lý kết nối Cluster",
      safe_mode_on: "Chế độ An toàn: BẬT",
      safe_mode_off: "Chế độ An toàn: TẮT",
      safe_mode_title_on: "Chế độ an toàn đang BẬT: Chặn các thao tác xóa hoặc cập nhật document ngoài ý muốn",
      safe_mode_title_off: "Chế độ an toàn đang TẮT: Cho phép thao tác ghi và xóa",
      opt_tips: "Mẹo tối ưu",
      opt_tips_title: "Nhấn để xem phân tích hiệu năng và tối ưu truy vấn",
      toggle_theme: "Chuyển giao diện Sáng / Tối",
      btn_format: "Làm đẹp",
      format_title: "Làm đẹp cú pháp JSON (Cmd+Shift+F)",
      btn_save: "Lưu mẫu",
      save_title: "Lưu truy vấn thành mẫu (Snippet)",
      btn_run: "Chạy truy vấn",
      run_title: "Thực thi truy vấn (Cmd+Enter hoặc Ctrl+Enter)",
      tab_indices: "Chỉ mục",
      tab_snippets: "Mẫu truy vấn",
      tab_history: "Lịch sử",
      filter_indices_ph: "Lọc tên chỉ mục...",
      refresh_indices: "Làm mới danh sách chỉ mục",
      loading_indices: "Đang tải danh sách chỉ mục...",
      no_indices_found: "Không tìm thấy chỉ mục nào",
      search_snippets_ph: "Tìm kiếm mẫu truy vấn...",
      no_snippets_yet: "Chưa có mẫu truy vấn nào",
      btn_clear_history: "Xóa toàn bộ lịch sử",
      no_history_yet: "Chưa có lịch sử truy vấn",
      target_index: "Chỉ mục:",
      target_index_title: "Chỉ mục mục tiêu cho tab truy vấn này",
      sync_line1: "Đồng bộ dòng 1",
      sync_line1_title: "Cập nhật endpoint 'POST /<index>/_search' vào dòng 1",
      new_tab: "Thêm Tab mới",
      json_valid: "JSON hợp lệ ✓",
      json_invalid: "Lỗi cú pháp JSON ✗",
      view_tree: "Cây JSON",
      view_table: "Dạng bảng",
      view_chart: "Biểu đồ Aggs",
      view_raw: "JSON thô",
      btn_copy: "Sao chép",
      copy_title: "Sao chép kết quả vào bộ nhớ tạm",
      export_csv: "Xuất CSV",
      search_tree_ph: "Tìm khóa hoặc giá trị trong JSON...",
      scope_hits: "Dữ liệu Hits (_source)",
      scope_hits_title: "Chỉ hiển thị dữ liệu tài liệu (_source) - gọn gàng, không rác metadata",
      scope_full: "Toàn bộ phản hồi",
      scope_full_title: "Hiển thị đầy đủ phong bì Elasticsearch (_shards, took, wrapper)",
      btn_expand_all: "Mở rộng hết",
      btn_collapse_all: "Thu gọn hết",
      empty_results: "Chạy truy vấn (Cmd+Enter) để xem kết quả tại đây",
      empty_table: "Không có dữ liệu hits. Hãy chạy truy vấn tìm kiếm có kết quả.",
      empty_chart: "Chạy truy vấn có chứa 'aggs' để xem biểu đồ trực quan tại đây",
      copied_clipboard: "Đã chép vào clipboard!",
      copied_path: "Đã chép đường dẫn JSON: ",
      doc_inspector_title: "Kiểm tra & Chỉnh sửa Tài liệu",
      btn_delete_doc: "Xóa tài liệu",
      btn_save_doc: "Lưu thay đổi",
      confirm_delete_doc: "Bạn có chắc chắn muốn xóa tài liệu này khỏi Elasticsearch?",
      safe_mode_blocked_delete: "Chế độ an toàn (Safe Mode) đang BẬT! Không thể xóa tài liệu.",
      safe_mode_blocked_write: "Chế độ an toàn (Safe Mode) đang BẬT! Không thể cập nhật tài liệu.",
      btn_ai_explain: "Giải thích AI",
      btn_ai_explain_title: "Giải thích truy vấn này bằng Antigravity",
      ai_prompt_ph: "Mô tả truy vấn bạn muốn (vd: tìm sản phẩm giá > 100k, sắp xếp mới nhất)...",
      btn_ai_generate: "Sinh Query",
      btn_ai_fix: "Tự sửa với Antigravity",
      ai_loading_generate: "Antigravity đang phân tích mapping và sinh Query DSL...",
      ai_loading_explain: "Antigravity đang phân tích và giải thích truy vấn...",
      ai_loading_fix: "Antigravity đang chẩn đoán lỗi và sửa truy vấn...",
      btn_golang: "Go Code",
      golang_modal_title: "Chuyển đổi Golang Code ⇋ Query DSL",
      tab_query_to_go: "Query DSL ➔ Golang Code",
      tab_go_to_query: "Golang Code ➔ Query DSL",
      btn_copy_go_code: "Sao chép Code Go",
      btn_copy_go_structs: "Chép Mẫu Structs Go",
      btn_compile_to_dsl: "Biên dịch sang Query DSL",
      btn_apply_dsl: "Áp dụng vào Trình soạn thảo"
    },
    en: {
      lang_flag: "🇬🇧",
      lang_label: "EN",
      btn_lang_title: "Switch to Vietnamese (Chuyển sang Tiếng Việt)",
      cluster_health: "Cluster Health",
      cluster_profiles: "Switch Cluster Profile",
      cluster_connecting: "Connecting...",
      manage_clusters: "Manage Cluster Connections",
      safe_mode_on: "Safe Mode: ON",
      safe_mode_off: "Safe Mode: OFF",
      safe_mode_title_on: "Safe Mode is ON: Blocks accidental document modifications or deletions",
      safe_mode_title_off: "Safe Mode is OFF: Modification and delete operations allowed",
      opt_tips: "Optimization Tips",
      opt_tips_title: "Click to view Query Performance & Optimization Tips",
      toggle_theme: "Toggle Light / Dark Theme",
      btn_format: "Format",
      format_title: "Format / Prettify Query (Cmd+Shift+F)",
      btn_save: "Save",
      save_title: "Save Query as Snippet",
      btn_run: "Run Query",
      run_title: "Execute Query (Cmd+Enter or Ctrl+Enter)",
      tab_indices: "Indices",
      tab_snippets: "Snippets",
      tab_history: "History",
      filter_indices_ph: "Filter indices...",
      refresh_indices: "Refresh Indices List",
      loading_indices: "Loading indices...",
      no_indices_found: "No indices found",
      search_snippets_ph: "Search snippets...",
      no_snippets_yet: "No saved snippets yet",
      btn_clear_history: "Clear History",
      no_history_yet: "No query history yet",
      target_index: "Target Index:",
      target_index_title: "Target index for this query tab",
      sync_line1: "Sync to Line 1",
      sync_line1_title: "Sync / update 'POST /<index>/_search' into Line 1",
      new_tab: "New Query Tab",
      json_valid: "JSON Valid ✓",
      json_invalid: "JSON Invalid ✗",
      view_tree: "JSON Tree",
      view_table: "Table View",
      view_chart: "Aggregations",
      view_raw: "Raw JSON",
      btn_copy: "Copy",
      copy_title: "Copy result to clipboard",
      export_csv: "Export CSV",
      search_tree_ph: "Search JSON keys or values...",
      scope_hits: "Hits Data (_source)",
      scope_hits_title: "Focus directly on hits/documents data (_source) - clean & no metadata clutter",
      scope_full: "Full Response",
      scope_full_title: "Show complete Elasticsearch response envelope (shards, took, raw wrapper)",
      btn_expand_all: "Expand All",
      btn_collapse_all: "Collapse All",
      empty_results: "Run a query (Cmd+Enter) to view results here",
      empty_table: "No table hits to display. Run a search query with hits.",
      empty_chart: "Run a query containing 'aggs' to view interactive charts here",
      copied_clipboard: "Copied to clipboard!",
      copied_path: "JSON Path copied: ",
      doc_inspector_title: "Document Inspector & Editor",
      btn_delete_doc: "Delete Document",
      btn_save_doc: "Save Changes",
      confirm_delete_doc: "Are you sure you want to delete this document from Elasticsearch?",
      safe_mode_blocked_delete: "Safe Mode is ON! Cannot delete documents.",
      safe_mode_blocked_write: "Safe Mode is ON! Cannot modify documents.",
      btn_ai_explain: "AI Explain",
      btn_ai_explain_title: "Explain this query using Antigravity",
      ai_prompt_ph: "Describe your query in natural language (e.g., find items price > 100, sort by created_at)...",
      btn_ai_generate: "Generate",
      btn_ai_fix: "Auto-Fix with Antigravity",
      ai_loading_generate: "Antigravity is analyzing schema and generating Query DSL...",
      ai_loading_explain: "Antigravity is analyzing and explaining query...",
      ai_loading_fix: "Antigravity is diagnosing error and fixing query...",
      btn_golang: "Go Code",
      golang_modal_title: "Convert Golang Code ⇋ Query DSL",
      tab_query_to_go: "Query DSL ➔ Golang Code",
      tab_go_to_query: "Golang Code ➔ Query DSL",
      btn_copy_go_code: "Copy Go Code",
      btn_copy_go_structs: "Copy Go Structs",
      btn_compile_to_dsl: "Compile to Query DSL",
      btn_apply_dsl: "Apply to Editor"
    }
  };

  function applyLanguage(lang) {
    if (!I18N[lang]) lang = 'vi';
    state.currentLang = lang;
    localStorage.setItem('eskhan_lang', lang);
    const dict = I18N[lang];

    // Update Flag & Label in header
    if (el.langFlag) el.langFlag.textContent = dict.lang_flag;
    if (el.langLabel) el.langLabel.textContent = dict.lang_label;
    if (el.btnToggleLang) el.btnToggleLang.title = dict.btn_lang_title;

    // Apply data-i18n text content
    document.querySelectorAll('[data-i18n]').forEach(elem => {
      const key = elem.getAttribute('data-i18n');
      if (dict[key]) {
        elem.textContent = dict[key];
      }
    });

    // Apply data-i18n-ph placeholders
    document.querySelectorAll('[data-i18n-ph]').forEach(elem => {
      const key = elem.getAttribute('data-i18n-ph');
      if (dict[key]) {
        elem.placeholder = dict[key];
      }
    });

    // Update Safe Mode button label & title
    if (el.safeModeLabel) {
      el.safeModeLabel.textContent = state.safeMode ? dict.safe_mode_on : dict.safe_mode_off;
    }
    if (el.btnSafeMode) {
      el.btnSafeMode.title = state.safeMode ? dict.safe_mode_title_on : dict.safe_mode_title_off;
    }
  }

  function toggleLanguage() {
    const nextLang = state.currentLang === 'vi' ? 'en' : 'vi';
    applyLanguage(nextLang);
  }

  // --- Initialization ---
  async function init() {
    applyTheme(state.theme);
    applyLanguage(state.currentLang);
    setupEventListeners();
    setupResizer();
    setupSidebarTabs();
    setupViewTabs();
    initMonaco();

    await loadClusters();
    await loadDSLSnippets();
    await loadSnippets();
    await loadHistory();
    await checkAntigravityStatus();
  }

  // --- Theme Management ---
  function applyTheme(theme) {
    state.theme = theme;
    if (theme === 'light') {
      document.body.classList.remove('theme-dark');
      document.body.classList.add('theme-light');
      if (el.themeIconSun) el.themeIconSun.classList.add('hidden');
      if (el.themeIconMoon) el.themeIconMoon.classList.remove('hidden');
      if (state.isMonacoLoaded && window.monaco) {
        monaco.editor.setTheme('vs');
      }
    } else {
      document.body.classList.remove('theme-light');
      document.body.classList.add('theme-dark');
      if (el.themeIconSun) el.themeIconSun.classList.remove('hidden');
      if (el.themeIconMoon) el.themeIconMoon.classList.add('hidden');
      if (state.isMonacoLoaded && window.monaco) {
        monaco.editor.setTheme('vs-dark');
      }
    }
    localStorage.setItem('eskhan_theme', theme);
  }

  function toggleTheme() {
    const nextTheme = state.theme === 'light' ? 'dark' : 'light';
    applyTheme(nextTheme);
  }

  // --- Monaco Editor Setup ---
  function initMonaco() {
    if (typeof require !== 'undefined') {
      require.config({
        paths: {
          vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.45.0/min/vs'
        }
      });

      require(['vs/editor/editor.main'], function () {
        state.isMonacoLoaded = true;
        createMonacoInstance();
        registerMonacoCompletionProviders();
        if (state.theme === 'light') {
          monaco.editor.setTheme('vs');
        }
      }, function (err) {
        console.warn('Monaco CDN failed to load, falling back to basic editor:', err);
        createFallbackEditor();
      });
    } else {
      createFallbackEditor();
    }
  }

  function createMonacoInstance() {
    el.editorContainer.innerHTML = '';
    const initialContent = state.tabs[0].content;

    state.editorInstance = monaco.editor.create(el.editorContainer, {
      value: initialContent,
      language: 'json',
      theme: state.theme === 'light' ? 'vs' : 'vs-dark',
      automaticLayout: true,
      fontSize: 13,
      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, monospace',
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      tabSize: 2,
      formatOnPaste: true,
      lineNumbers: 'on',
      folding: true,
      tabCompletion: 'on',
      quickSuggestions: { other: true, comments: false, strings: true },
      suggestOnTriggerCharacters: true,
      acceptSuggestionOnEnter: 'on',
      acceptSuggestionOnCommitCharacter: true,
      snippetSuggestions: 'top',
      wordBasedSuggestions: 'off',
      suggest: {
        preview: true,
        showWords: false,
        showSnippets: true,
        insertMode: 'insert'
      }
    });

    state.editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, function () {
      runCurrentQuery();
    });

    state.editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyF, function () {
      formatQueryText();
    });

    state.editorInstance.onDidChangeModelContent(function () {
      const activeTab = state.tabs.find(t => t.id === state.activeTabId);
      if (activeTab) {
        activeTab.content = state.editorInstance.getValue();
      }
      validateJSON(activeTab.content);
      triggerLinterDebounced(activeTab.content);
    });

    triggerLinterDebounced(initialContent);
  }

  function createFallbackEditor() {
    el.editorContainer.innerHTML = `
      <textarea id="fallbackTextarea" class="form-input" style="width: 100%; height: 100%; resize: none; font-family: var(--font-mono); font-size: 13px; background: #1e1e1e; color: #f4f4f5; border: none; padding: 12px; outline: none;">${state.tabs[0].content}</textarea>
    `;
    const textarea = document.getElementById('fallbackTextarea');

    textarea.addEventListener('keydown', function (e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        runCurrentQuery();
      }
      if (e.key === 'Tab') {
        e.preventDefault();
        const start = this.selectionStart;
        const end = this.selectionEnd;
        this.value = this.value.substring(0, start) + '  ' + this.value.substring(end);
        this.selectionStart = this.selectionEnd = start + 2;
      }
    });

    textarea.addEventListener('input', function () {
      const activeTab = state.tabs.find(t => t.id === state.activeTabId);
      if (activeTab) {
        activeTab.content = this.value;
      }
      validateJSON(this.value);
      triggerLinterDebounced(this.value);
    });

    state.editorInstance = {
      getValue: () => textarea.value,
      setValue: (val) => { textarea.value = val; },
      focus: () => textarea.focus()
    };
  }

  // --- Context-Aware Autocomplete Provider ---
  function registerMonacoCompletionProviders() {
    monaco.languages.registerCompletionItemProvider('json', {
      triggerCharacters: ['"', ':', '{', ' ', '.', '$', '_', '/', 'q', 'm', 'b', 'f', 's', 'r', 't', 'a', 'p', 'g', 'd', 'e', 'w', 'v', 'i', 'h', 'l', 'k', 'c'],
      provideCompletionItems: async function (model, position) {
        const lineContent = model.getLineContent(position.lineNumber);
        const col = position.column;
        const lineContentBefore = lineContent.substring(0, col - 1);
        const textBefore = model.getValueInRange({
          startLineNumber: Math.max(1, position.lineNumber - 5),
          startColumn: 1,
          endLineNumber: position.lineNumber,
          endColumn: position.column
        });

        const word = model.getWordUntilPosition(position);
        let startCol = word.startColumn;
        let endCol = word.endColumn;

        const hasOpenQuote = (startCol > 1 && lineContent.charAt(startCol - 2) === '"');
        const hasCloseQuote = (endCol <= lineContent.length && lineContent.charAt(endCol - 1) === '"');

        const replaceRange = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: hasOpenQuote ? startCol - 1 : startCol,
          endColumn: hasCloseQuote ? endCol + 1 : endCol
        };

        const bareRange = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: startCol,
          endColumn: endCol
        };

        const suggestions = [];

        // 1. Line 1: HTTP Method & Index Endpoint Suggestions
        if (position.lineNumber === 1 && !lineContent.trim().startsWith('{')) {
          state.indices.forEach(idx => {
            suggestions.push({
              label: `POST /${idx.name}/_search`,
              kind: monaco.languages.CompletionItemKind.Method,
              insertText: `POST /${idx.name}/_search\n{\n  "size": 20,\n  "query": {\n    $0\n  }\n}`,
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              detail: `Query ${idx.name} (${formatNumber(idx.docs_count)} docs)`,
              sortText: `0_${idx.name}`,
              range: bareRange
            });
            suggestions.push({
              label: `GET /${idx.name}/_mapping`,
              kind: monaco.languages.CompletionItemKind.Method,
              insertText: `GET /${idx.name}/_mapping`,
              detail: `Inspect mapping for ${idx.name}`,
              sortText: `1_${idx.name}`,
              range: bareRange
            });
            suggestions.push({
              label: `/${idx.name}/_search`,
              kind: monaco.languages.CompletionItemKind.Module,
              insertText: `/${idx.name}/_search`,
              detail: `Endpoint for ${idx.name}`,
              sortText: `2_${idx.name}`,
              range: bareRange
            });
          });
        }

        // 2. DSL Keywords & Clauses with Tab-stop Snippets
        const dslKeywords = [
          // Vietnamese Specialized Search Snippets
          {
            label: 'tiengviet_search',
            detail: 'Tìm kiếm tiếng Việt có dấu/không dấu (multi_match + highlight)',
            insertText: '"query": {\n  "multi_match": {\n    "query": "${1:điện thoại thông minh}",\n    "fields": ["${2:name^3}", "${3:title^2}", "${4:description}"],\n    "type": "best_fields",\n    "operator": "and"\n  }\n},\n"highlight": {\n  "pre_tags": ["<mark>"],\n  "post_tags": ["</mark>"],\n  "fields": {\n    "${2:name}": {},\n    "${4:description}": {}\n  }\n}'
          },
          {
            label: 'tiengviet_goiy',
            detail: 'Gợi ý từ khóa tiếng Việt (match_phrase_prefix autocomplete)',
            insertText: '"query": {\n  "match_phrase_prefix": {\n    "${1:name}": {\n      "query": "${2:máy giặt}",\n      "max_expansions": 10\n    }\n  }\n}'
          },
          {
            label: 'tiengviet_boloc',
            detail: 'Tìm kiếm tiếng Việt kết hợp bộ lọc bool (must + filter term/range)',
            insertText: '"query": {\n  "bool": {\n    "must": [\n      {\n        "multi_match": {\n          "query": "${1:áo sơ mi}",\n          "fields": ["${2:title^2}", "${3:content}"],\n          "operator": "and"\n        }\n      }\n    ],\n    "filter": [\n      { "term": { "${4:status}": "${5:active}" } },\n      { "range": { "${6:price}": { "gte": 100000, "lte": 500000 } } }\n    ]\n  }\n}'
          },
          // Standard DSL Clauses
          { label: 'query', detail: 'ES Root query clause', insertText: '"query": {\n  $0\n}' },
          { label: 'bool', detail: 'Compound boolean query', insertText: '"bool": {\n  "must": [\n    $0\n  ]\n}' },
          { label: 'must', detail: 'Must match clause (AND)', insertText: '"must": [\n  $0\n]' },
          { label: 'filter', detail: 'Filter clause (cached, no score)', insertText: '"filter": [\n  $0\n]' },
          { label: 'should', detail: 'Should match clause (OR)', insertText: '"should": [\n  $0\n]' },
          { label: 'must_not', detail: 'Must not match (NOT)', insertText: '"must_not": [\n  $0\n]' },
          { label: 'match', detail: 'Full-text search query', insertText: '"match": {\n  "${1:field}": "${2:text}"\n}' },
          { label: 'match_phrase', detail: 'Exact phrase query with word ordering (từ ghép)', insertText: '"match_phrase": {\n  "${1:field}": {\n    "query": "${2:cụm từ tiếng Việt}",\n    "slop": ${3:1}\n  }\n}' },
          { label: 'match_phrase_prefix', detail: 'Prefix phrase search (autocomplete/search-as-you-type)', insertText: '"match_phrase_prefix": {\n  "${1:field}": {\n    "query": "${2:tiền tố}",\n    "max_expansions": 10\n  }\n}' },
          { label: 'multi_match', detail: 'Multi-field search query with field weights', insertText: '"multi_match": {\n  "query": "${1:query text}",\n  "fields": ["${2:name^3}", "${3:description}"],\n  "type": "${4|best_fields,most_fields,cross_fields,phrase|}",\n  "operator": "${5|and,or|}"\n}' },
          { label: 'match_all', detail: 'Match all documents', insertText: '"match_all": {}' },
          { label: 'term', detail: 'Exact term filter', insertText: '"term": {\n  "${1:field}": "${2:value}"\n}' },
          { label: 'terms', detail: 'Exact terms multi-filter', insertText: '"terms": {\n  "${1:field}": ["${2:val1}", "${3:val2}"]\n}' },
          { label: 'range', detail: 'Numeric or date range', insertText: '"range": {\n  "${1:field}": {\n    "gte": ${2:0},\n    "lte": ${3:100}\n  }\n}' },
          { label: 'exists', detail: 'Document contains field', insertText: '"exists": {\n  "field": "${1:field}"\n}' },
          { label: 'wildcard', detail: 'Wildcard pattern query', insertText: '"wildcard": {\n  "${1:field}": "${2:*value*}"\n}' },
          { label: 'prefix', detail: 'Prefix term query', insertText: '"prefix": {\n  "${1:field}": "${2:prefix}"\n}' },
          { label: 'operator', detail: 'Match boolean operator (and / or)', insertText: '"operator": "${1|and,or|}"' },
          { label: 'fuzziness', detail: 'Fuzzy edit distance for typos', insertText: '"fuzziness": "${1|AUTO,0,1,2|}"' },
          { label: 'tie_breaker', detail: 'Tie breaker factor for multi_match (0.0 to 1.0)', insertText: '"tie_breaker": ${1:0.3}' },
          { label: 'aggs', detail: 'Aggregations clause', insertText: '"aggs": {\n  "${1:agg_name}": {\n    "terms": {\n      "field": "${2:field.keyword}",\n      "size": ${3:10}\n    }\n  }\n}' },
          { label: 'size', detail: 'Number of hits to return', insertText: '"size": ${1:20}' },
          { label: 'from', detail: 'Offset pagination', insertText: '"from": ${1:0}' },
          { label: 'sort', detail: 'Sort order', insertText: '"sort": [\n  { "${1:field}": { "order": "${2|desc,asc|}" } }\n]' },
          { label: 'highlight', detail: 'Highlight matches with <mark> tags', insertText: '"highlight": {\n  "pre_tags": ["<mark>"],\n  "post_tags": ["</mark>"],\n  "fields": {\n    "${1:field}": {}\n  }\n}' },
          { label: '_source', detail: 'Filter returned fields', insertText: '"_source": ["${1:field1}", "${2:field2}"]' },
          { label: 'track_total_hits', detail: 'Accurate total count (>10k)', insertText: '"track_total_hits": true' }
        ];

        // 3. System Variables
        const sysVars = [
          { label: '{{$timestamp}}', detail: 'Current Unix timestamp in seconds', insertText: '{{$timestamp}}' },
          { label: '{{$timestamp_ms}}', detail: 'Current Unix timestamp in ms', insertText: '{{$timestamp_ms}}' },
          { label: '{{$date}}', detail: 'Current date YYYY-MM-DD', insertText: '{{$date}}' },
          { label: '{{$uuid}}', detail: 'Generated unique UUID', insertText: '{{$uuid}}' }
        ];

        sysVars.forEach(v => {
          suggestions.push({
            label: v.label,
            kind: monaco.languages.CompletionItemKind.Variable,
            insertText: v.insertText,
            detail: v.detail,
            range: bareRange
          });
        });

        dslKeywords.forEach(k => {
          suggestions.push({
            label: k.label,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: k.insertText,
            insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
            detail: k.detail,
            range: replaceRange
          });
        });

        // 4. Index names inside JSON (e.g. "index": "...")
        if (lineContentBefore.includes('"index"')) {
          state.indices.forEach(idx => {
            suggestions.push({
              label: `"${idx.name}"`,
              kind: monaco.languages.CompletionItemKind.Value,
              insertText: `"${idx.name}"`,
              detail: `Index (${formatNumber(idx.docs_count)} docs)`,
              range: replaceRange
            });
          });
        }

        // 5. Context-Aware Dynamic Fields from mapping
        const activeIndex = state.activeIdx;
        if (activeIndex) {
          const fields = await fetchIndexFields(activeIndex);

          // Context detection
          const isRangeContext = /"range"\s*:\s*\{[^}]*$/i.test(textBefore);
          const isMatchContext = /"match"\s*:\s*\{[^}]*$/i.test(textBefore);
          const isTermsContext = /"terms?"\s*:\s*\{[^}]*$/i.test(textBefore) || /"aggs?"\s*:\s*\{[^}]*$/i.test(textBefore);

          fields.forEach(f => {
            let priority = 0;
            if (isRangeContext && (f.type === 'date' || f.type === 'long' || f.type === 'integer' || f.type === 'float')) {
              priority = 10;
            } else if (isMatchContext && (f.type === 'text')) {
              priority = 10;
            } else if (isTermsContext && (f.type === 'keyword' || f.type === 'integer' || f.type === 'long')) {
              priority = 10;
            }

            suggestions.push({
              label: f.name,
              kind: monaco.languages.CompletionItemKind.Field,
              insertText: `"${f.name}"`,
              detail: `[${f.type}] ${f.detail || 'Field'}`,
              sortText: priority > 0 ? `00_${f.name}` : `01_${f.name}`,
              range: replaceRange
            });
          });

          // Top distinct terms suggestions
          if (lineContentBefore.includes(':')) {
            const match = lineContentBefore.match(/"([^"]+)"\s*:\s*"?$/);
            if (match && match[1]) {
              const fieldName = match[1];
              const fieldObj = fields.find(f => f.name === fieldName);
              if (fieldObj && fieldObj.type === 'keyword') {
                const terms = await fetchTopTerms(activeIndex, fieldName);
                terms.forEach(t => {
                  suggestions.push({
                    label: `"${t}"`,
                    kind: monaco.languages.CompletionItemKind.Value,
                    insertText: `"${t}"`,
                    detail: `Top term for ${fieldName}`,
                    sortText: `000_${t}`,
                    range: replaceRange
                  });
                });
              }
            }
          }
        }

        return { suggestions: suggestions };
      }
    });
  }

  // --- Smart Query Linter ---

  function triggerLinterDebounced(queryText) {
    clearTimeout(state.linterTimeout);
    state.linterTimeout = setTimeout(() => {
      runQueryLinter(queryText);
    }, 400);
  }

  async function runQueryLinter(queryText) {
    if (!queryText || !queryText.trim()) {
      el.linterBadge.classList.add('hidden');
      return;
    }

    try {
      const res = await apiPost('/api/linter/check', {
        query: queryText,
        index: state.activeIdx
      });
      state.linterWarnings = res.warnings || [];
      if (state.linterWarnings.length > 0) {
        el.linterCount.textContent = state.linterWarnings.length;
        el.linterBadge.classList.remove('hidden');
      } else {
        el.linterBadge.classList.add('hidden');
      }
    } catch (e) {
      // Silent error for linter
    }
  }

  function showLinterModal() {
    if (state.linterWarnings.length === 0) return;
    el.linterModal.classList.add('active');
    el.linterWarningsList.innerHTML = state.linterWarnings.map(w => `
      <div class="linter-warning-card ${w.severity}">
        <div class="linter-warning-title"><span class="tag-badge" style="color:var(--warning); margin-right:4px;">${w.severity.toUpperCase()}</span> ${escapeHtml(w.rule_id)}</div>
        <div class="linter-warning-msg">${escapeHtml(w.message)}</div>
        ${w.suggestion ? `<div class="linter-warning-suggestion"><strong>Suggestion:</strong> ${escapeHtml(w.suggestion)}</div>` : ''}
      </div>
    `).join('');
  }

  // --- API Calls ---

  async function apiGet(endpoint) {
    const res = await fetch(endpoint);
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }
    return res.json();
  }

  async function apiPost(endpoint, data) {
    const res = await fetch(endpoint, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }
    return res.json();
  }

  async function apiDelete(endpoint) {
    const res = await fetch(endpoint, { method: 'DELETE' });
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }
    return res.json();
  }

  // --- Clusters & Connection Profile ---

  async function loadClusters() {
    try {
      const data = await apiGet('/api/connections');
      state.clusters = data.connections || [];
      const activeId = data.active_id;

      el.clusterSelect.innerHTML = '';
      state.clusters.forEach(c => {
        const opt = document.createElement('option');
        opt.value = c.id;
        opt.textContent = `${c.name} (${c.url})`;
        if (c.id === activeId) opt.selected = true;
        el.clusterSelect.appendChild(opt);
      });

      state.activeCluster = state.clusters.find(c => c.id === activeId) || state.clusters[0];
      await checkClusterHealth();
      await loadIndices();
    } catch (err) {
      setClusterBadge('gray', 'No Connection');
      console.error('Failed to load clusters:', err);
    }
  }

  async function checkClusterHealth() {
    try {
      setClusterBadge('gray', 'Pinging...');
      const data = await apiGet('/api/cluster/info');
      const health = data.health ? data.health.status : 'green';
      const version = data.info && data.info.version ? data.info.version.number : 'ES';
      setClusterBadge(health, `ES ${version} (${health})`);
    } catch (err) {
      setClusterBadge('red', 'Cluster Offline');
    }
  }

  function setClusterBadge(color, text) {
    el.clusterStatusBadge.querySelector('.status-dot').className = `status-dot dot-${color}`;
    el.clusterStatusText.textContent = text;
  }

  // --- Indices Management & Operations ---

  async function loadIndices() {
    el.indicesList.innerHTML = '<div class="empty-state">Loading indices...</div>';
    try {
      const data = await apiGet('/api/indices');
      state.indices = data.indices || [];
      renderIndices(state.indices);
      populateIndexDropdown(state.indices);
    } catch (err) {
      el.indicesList.innerHTML = `<div class="empty-state" style="color:var(--danger)">Error: ${err.message}</div>`;
    }
  }

  function populateIndexDropdown(indices) {
    const opts = ['<option value="">(All Indices / _search)</option>'];
    indices.forEach(idx => {
      opts.push(`<option value="${escapeHtml(idx.name)}">${escapeHtml(idx.name)} (${formatNumber(idx.docs_count)} docs)</option>`);
    });
    const html = opts.join('');
    if (el.editorIndexSelect) {
      el.editorIndexSelect.innerHTML = html;
      el.editorIndexSelect.value = state.activeIdx;
    }
  }

  function renderIndices(indices) {
    const filter = el.indicesSearchInput.value.toLowerCase().trim();
    const filtered = indices.filter(i => i.name.toLowerCase().includes(filter));

    if (filtered.length === 0) {
      el.indicesList.innerHTML = '<div class="empty-state">No indices found</div>';
      return;
    }

    el.indicesList.innerHTML = '';
    filtered.forEach(idx => {
      const card = document.createElement('div');
      card.className = `index-card ${idx.name === state.activeIdx ? 'active' : ''}`;
      card.innerHTML = `
        <div class="index-card-title">
          <span><span class="status-dot dot-${idx.health}"></span> ${escapeHtml(idx.name)}</span>
          <span style="font-size:10px; color:var(--text-muted)">${idx.store_size || '-'}</span>
        </div>
        <div class="index-card-meta">
          <span>Docs: ${formatNumber(idx.docs_count)}</span>
          <span>Shards: ${idx.pri}p/${idx.rep}r</span>
        </div>
        <div class="index-actions">
          <button class="btn btn-sm btn-primary btn-select-idx" data-idx="${escapeHtml(idx.name)}" title="Set as target index for current query">Select</button>
          <button class="btn btn-sm btn-secondary btn-new-tab-idx" data-idx="${escapeHtml(idx.name)}" title="Open new tab for this index">+ Tab</button>
          <button class="btn btn-sm btn-secondary btn-inspect" data-idx="${escapeHtml(idx.name)}" title="Inspect Mapping Schema">Schema</button>
          <button class="index-ops-btn btn-refresh-idx" title="Refresh Index">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
          </button>
          <button class="index-ops-btn btn-clear-cache" title="Clear Query Cache">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"></path></svg>
          </button>
        </div>
      `;

      card.querySelector('.btn-select-idx').addEventListener('click', (e) => {
        e.stopPropagation();
        setTargetIndex(idx.name, true);
      });

      card.querySelector('.btn-new-tab-idx').addEventListener('click', (e) => {
        e.stopPropagation();
        createNewTabWithIndex(idx.name);
      });

      card.querySelector('.btn-inspect').addEventListener('click', (e) => {
        e.stopPropagation();
        inspectIndexMapping(idx.name);
      });

      card.querySelector('.btn-refresh-idx').addEventListener('click', async (e) => {
        e.stopPropagation();
        await triggerIndexAction(idx.name, 'refresh');
      });

      card.querySelector('.btn-clear-cache').addEventListener('click', async (e) => {
        e.stopPropagation();
        await triggerIndexAction(idx.name, 'cache_clear');
      });

      card.addEventListener('click', () => {
        setTargetIndex(idx.name, true);
      });

      el.indicesList.appendChild(card);
    });
  }

  function createNewTabWithIndex(idxName) {
    const newId = state.tabs.length > 0 ? Math.max(...state.tabs.map(t => t.id)) + 1 : 1;
    const newTab = {
      id: newId,
      title: `${idxName}`,
      index: idxName,
      content: `POST /${idxName}/_search\n{\n  "size": 20,\n  "query": {\n    "match_all": {}\n  }\n}`
    };
    state.tabs.push(newTab);
    renderEditorTabs();
    switchEditorTab(newId);
  }

  async function triggerIndexAction(indexName, action) {
    if (state.safeMode && action !== 'refresh' && action !== 'cache_clear') {
      alert('Safe Mode is ON. Action restricted.');
      return;
    }
    try {
      await apiPost(`/api/indices/${encodeURIComponent(indexName)}/action`, { action });
      alert(`✓ ${action.toUpperCase()} executed successfully on index "${indexName}"`);
      await loadIndices();
    } catch (err) {
      alert(`Action failed: ${err.message}`);
    }
  }

  function setTargetIndex(idxName, syncEditor = true) {
    state.activeIdx = idxName;
    if (el.editorIndexSelect) el.editorIndexSelect.value = idxName;

    const activeTab = state.tabs.find(t => t.id === state.activeTabId);
    if (activeTab) {
      activeTab.index = idxName;
    }

    // Automatically update endpoint line 1 if present
    if (syncEditor && state.editorInstance && idxName) {
      const val = state.editorInstance.getValue();
      const lines = val.split('\n');
      const firstLine = lines[0].trim();
      const match = firstLine.match(/^([A-Z]+\s+)?\/([^/\s]+)(\/.*)?$/);
      if (match) {
        const method = match[1] || 'POST ';
        const rest = match[3] || '/_search';
        lines[0] = `${method}/${idxName}${rest}`;
        const newContent = lines.join('\n');
        state.editorInstance.setValue(newContent);
        if (activeTab) activeTab.content = newContent;
      }
    }

    renderIndices(state.indices);
    if (state.editorInstance) {
      triggerLinterDebounced(state.editorInstance.getValue());
    }
  }

  function setActiveIndex(idxName) {
    setTargetIndex(idxName, true);
  }

  function insertEndpointToLine1(idxName) {
    const target = idxName || state.activeIdx;
    if (!state.editorInstance) return;
    const val = state.editorInstance.getValue();
    const lines = val.split('\n');
    const firstLine = lines[0].trim();
    const endpoint = `POST /${target || '_search'}/_search`;

    const match = firstLine.match(/^([A-Z]+\s+)?\/([^/\s]+)(\/.*)?$/);
    if (match) {
      lines[0] = endpoint;
      state.editorInstance.setValue(lines.join('\n'));
    } else {
      state.editorInstance.setValue(`${endpoint}\n${val}`);
    }
    const activeTab = state.tabs.find(t => t.id === state.activeTabId);
    if (activeTab) activeTab.content = state.editorInstance.getValue();
  }

  function selectIndexAndTemplate(idxName) {
    setTargetIndex(idxName, true);
    const query = `POST /${idxName}/_search\n{\n  "size": 20,\n  "query": {\n    "match_all": {}\n  }\n}`;
    state.editorInstance.setValue(query);
    runCurrentQuery();
  }

  async function fetchIndexFields(indexName) {
    if (state.mappingCache[indexName]) {
      return state.mappingCache[indexName];
    }
    try {
      const data = await apiGet(`/api/indices/${encodeURIComponent(indexName)}/mapping`);
      state.mappingCache[indexName] = data.fields || [];
      return state.mappingCache[indexName];
    } catch (e) {
      return [];
    }
  }

  async function fetchTopTerms(indexName, field) {
    const cacheKey = `${indexName}:${field}`;
    if (state.topTermsCache[cacheKey]) {
      return state.topTermsCache[cacheKey];
    }
    try {
      const data = await apiGet(`/api/indices/${encodeURIComponent(indexName)}/terms?field=${encodeURIComponent(field)}&size=8`);
      state.topTermsCache[cacheKey] = data.terms || [];
      return state.topTermsCache[cacheKey];
    } catch (e) {
      return [];
    }
  }

  async function inspectIndexMapping(indexName) {
    el.mappingModalTitle.textContent = `Index Mapping: ${indexName}`;
    el.mappingModal.classList.add('active');
    el.mappingFieldsBody.innerHTML = '<tr><td colspan="3" style="text-align:center; padding:16px;">Loading fields...</td></tr>';

    const fields = await fetchIndexFields(indexName);
    renderMappingTable(fields);

    el.mappingFilterInput.oninput = () => {
      const filter = el.mappingFilterInput.value.toLowerCase().trim();
      const filtered = fields.filter(f => f.name.toLowerCase().includes(filter) || f.type.toLowerCase().includes(filter));
      renderMappingTable(filtered);
    };

    el.btnInsertTemplateQuery.onclick = () => {
      el.mappingModal.classList.remove('active');
      selectIndexAndTemplate(indexName);
    };
  }

  function renderMappingTable(fields) {
    if (fields.length === 0) {
      el.mappingFieldsBody.innerHTML = '<tr><td colspan="3" style="text-align:center; padding:16px;">No fields found</td></tr>';
      return;
    }
    el.mappingFieldsBody.innerHTML = fields.map(f => `
      <tr>
        <td style="color:var(--accent); font-weight:500;">${escapeHtml(f.name)}</td>
        <td><span class="tag-badge">[${escapeHtml(f.type)}]</span></td>
        <td style="color:var(--text-muted)">${escapeHtml(f.detail || '')}</td>
      </tr>
    `).join('');
  }

  // --- Query Execution & Results Rendering ---

  async function runCurrentQuery() {
    const rawContent = state.editorInstance.getValue().trim();
    if (!rawContent) return;

    el.btnRunQuery.disabled = true;
    el.btnRunQuery.innerHTML = `
      <svg class="icon icon-sm" style="animation: spin 1s linear infinite;" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10" stroke-dasharray="32" stroke-linecap="round"></circle></svg>
      <span>Executing...</span>
    `;
    setMetricStatus('loading', 'Executing...');

    try {
      const payload = {
        raw_input: rawContent,
        index: state.activeIdx
      };

      const result = await apiPost('/api/query/execute', payload);
      state.lastResult = result;
      displayResults(result);
      await loadHistory();
    } catch (err) {
      displayErrorResult(err.message);
    } finally {
      el.btnRunQuery.disabled = false;
      el.btnRunQuery.innerHTML = `
        <svg class="icon icon-play" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
        <span>Run Query</span>
        <kbd class="shortcut-badge">⌘⏎</kbd>
      `;
    }
  }

  function displayResults(res) {
    const isSuccess = res.status_code >= 200 && res.status_code < 300;
    setMetricStatus(isSuccess ? 'success' : 'error', `HTTP ${res.status_code} ${res.status_text}`);

    el.metricTook.textContent = `Took: ${res.took_ms} ms`;
    el.metricLatency.textContent = `Latency: ${res.latency_ms} ms`;
    el.metricHits.textContent = `Hits: ${formatNumber(res.total_hits)} (${res.total_hits_relation || 'eq'})`;
    el.metricShards.textContent = `Shards: ${res.shards.successful}/${res.shards.total}`;

    // Render Raw JSON
    el.rawJsonViewer.textContent = formatRawJSON(res.raw_json);

    // Render Smart JSON Tree (Hits Data or Full Envelope)
    renderJsonTree();

    if (!isSuccess) {
      const banner = document.createElement('div');
      banner.className = 'ai-error-banner';
      banner.style.cssText = 'background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); border-radius: 6px; padding: 10px 14px; margin-bottom: 12px; display: flex; align-items: center; justify-content: space-between; gap: 12px;';
      banner.innerHTML = `
        <div style="color: var(--danger); font-size: 13px; font-weight: 500; display:flex; align-items:center; gap:8px;">
          <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          <span>Lỗi Elasticsearch (${res.status_code}): Bạn có muốn Antigravity tự động sửa không?</span>
        </div>
        <button class="btn btn-sm btn-primary" id="btnAIFixResultError" style="gap:6px; white-space:nowrap;">
          <svg class="icon icon-sm ai-sparkle-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2l2.4 5.6L20 10l-4.4 4 1.4 6-5-3.2-5 3.2 1.4-6-4.4-4 5.6-2.4z"></path></svg>
          <span>${state.currentLang === 'vi' ? 'Tự sửa bằng Antigravity' : 'Auto-Fix with Antigravity'}</span>
        </button>
      `;
      el.jsonTreeContainer.prepend(banner);
      const btn = document.getElementById('btnAIFixResultError');
      if (btn) btn.addEventListener('click', () => fixCurrentQueryWithError(res.raw_json));
    }

    // Render Table View with Edit/Inspect actions
    renderHitsTable(res.extracted_hits, res.hit_columns);

    // Render Aggregations Charts if present
    if (res.aggregations && Object.keys(res.aggregations).length > 0) {
      el.tabViewChart.style.display = 'inline-flex';
      renderAggregationsCharts(res.aggregations);
    } else {
      el.tabViewChart.style.display = 'none';
      if (state.currentView === 'chart') {
        switchResultView('tree');
      }
    }
  }

  function renderJsonTree() {
    if (!state.lastResult || !state.lastResult.raw_json) return;
    try {
      const parsed = JSON.parse(state.lastResult.raw_json);
      el.jsonTreeContainer.innerHTML = '';

      const isSearchHits = parsed.hits && Array.isArray(parsed.hits.hits);

      if (state.jsonScope === 'hits' && isSearchHits) {
        const hits = parsed.hits.hits;

        // Render Aggregations block if present
        if (parsed.aggregations && Object.keys(parsed.aggregations).length > 0) {
          el.jsonTreeContainer.appendChild(createJsonTreeNode(parsed.aggregations, 'aggregations', '$', true));
        }

        if (hits.length === 0) {
          const empty = document.createElement('div');
          empty.className = 'empty-state';
          empty.innerHTML = 'No matching documents found in hits (total: 0)';
          el.jsonTreeContainer.appendChild(empty);
          return;
        }

        // Summary bar
        const totalVal = parsed.hits.total ? (parsed.hits.total.value !== undefined ? parsed.hits.total.value : parsed.hits.total) : hits.length;
        const summary = document.createElement('div');
        summary.className = 'json-hits-summary';
        summary.innerHTML = `
          <span>Showing <strong>${hits.length}</strong> of <strong>${formatNumber(totalVal)}</strong> documents (Hits / _source Mode)</span>
          <span style="font-size:11px; color:var(--text-muted);">Envelope hidden. Click "Full Response" to view shards & took</span>
        `;
        el.jsonTreeContainer.appendChild(summary);

        // Render each document directly unwrapping _source
        hits.forEach((h, idx) => {
          const docId = h._id || idx + 1;
          const scoreStr = h._score !== undefined && h._score !== null ? ` | score: ${h._score}` : '';
          const label = `Doc #${idx + 1} [id: ${docId}${scoreStr}]`;
          
          let docData = h._source !== undefined ? { ...h._source } : h;
          if (h.highlight) {
            docData = { _highlight: h.highlight, ...docData };
          }
          
          // Auto-expand first 2 docs for immediate productivity
          const autoExpand = idx < 2;
          el.jsonTreeContainer.appendChild(createJsonTreeNode(docData, label, `$.hits[${idx}]._source`, autoExpand));
        });
      } else {
        // Full Response mode or non-search response
        el.jsonTreeContainer.appendChild(createJsonTreeNode(parsed, 'root', '$', true));
      }
    } catch (e) {
      el.jsonTreeContainer.innerHTML = `<div class="empty-state">${escapeHtml(state.lastResult.raw_json)}</div>`;
    }
  }

  function displayErrorResult(errMsg) {
    setMetricStatus('error', 'Execution Error');
    el.rawJsonViewer.textContent = errMsg;
    el.jsonTreeContainer.innerHTML = `
      <div class="empty-state" style="color:var(--danger); text-align:left; max-width:650px; margin:20px auto;">
        <div style="font-weight:600; margin-bottom:8px; font-size:14px;">⚠️ Lỗi thực thi truy vấn:</div>
        <div style="background:var(--bg-card); padding:12px; border-radius:6px; font-family:monospace; font-size:12px; white-space:pre-wrap; margin-bottom:14px; border:1px solid var(--border); line-height:1.5;">${escapeHtml(errMsg)}</div>
        <button class="btn btn-sm btn-primary" id="btnAIFixError" style="gap:6px;">
          <svg class="icon icon-sm ai-sparkle-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2l2.4 5.6L20 10l-4.4 4 1.4 6-5-3.2-5 3.2 1.4-6-4.4-4 5.6-2.4z"></path></svg>
          <span>${state.currentLang === 'vi' ? 'Tự sửa bằng Antigravity' : 'Auto-Fix with Antigravity'}</span>
        </button>
      </div>
    `;
    el.tableContainer.innerHTML = `<div class="empty-state" style="color:var(--danger)">${escapeHtml(errMsg)}</div>`;

    const btnFix = document.getElementById('btnAIFixError');
    if (btnFix) {
      btnFix.addEventListener('click', () => fixCurrentQueryWithError(errMsg));
    }
  }

  function setMetricStatus(type, text) {
    el.metricStatus.className = `metric-item status-badge ${type}`;
    el.metricStatus.textContent = text;
  }

  // --- High-Contrast, Professional JSON Tree View Generator ---
  function createJsonTreeNode(value, key, path = '$', autoExpand = false) {
    const nodeWrapper = document.createElement('div');
    nodeWrapper.className = 'json-node';
    nodeWrapper.setAttribute('data-key', String(key).toLowerCase());

    const isRoot = key === 'root';
    const currentPath = isRoot ? '$' : (Array.isArray(value) ? `${path}[${key}]` : `${path}.${key}`);

    if (value === null) {
      const row = document.createElement('div');
      row.className = 'json-row';
      row.innerHTML = `
        <span class="json-key">${escapeHtml(key)}</span><span class="json-colon">:</span>
        <span class="json-null">null</span>
        <div class="json-row-actions">
          <button class="btn-mini-copy" data-copy="null" title="Copy value">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
          </button>
          <button class="btn-mini-copy" data-copy-path="${escapeHtml(currentPath)}" title="Copy JSON path">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
          </button>
        </div>
      `;
      attachMiniCopy(row);
      nodeWrapper.appendChild(row);
      return nodeWrapper;
    }

    const type = typeof value;
    if (type === 'object') {
      const isArray = Array.isArray(value);
      const keys = Object.keys(value);
      const openBracket = isArray ? '[' : '{';
      const closeBracket = isArray ? ']' : '}';

      const row = document.createElement('div');
      row.className = 'json-row json-object-header';
      row.innerHTML = `
        <span class="json-toggle-chevron ${autoExpand ? '' : 'collapsed'}">
          <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"></polyline></svg>
        </span>
        ${!isRoot ? `<span class="json-key">${escapeHtml(key)}</span><span class="json-colon">:</span> ` : ''}
        <span class="json-bracket">${openBracket}</span>
        <span class="json-count-badge">${keys.length} ${keys.length === 1 ? 'item' : 'items'}</span>
        <div class="json-row-actions">
          <button class="btn-mini-copy" data-copy-json="true" title="Copy JSON object">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
          </button>
          <button class="btn-mini-copy" data-copy-path="${escapeHtml(currentPath)}" title="Copy JSON path">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
          </button>
        </div>
      `;

      const chevron = row.querySelector('.json-toggle-chevron');
      const childrenWrapper = document.createElement('div');
      childrenWrapper.className = 'json-children-wrapper';
      childrenWrapper.style.display = autoExpand ? 'block' : 'none';

      keys.forEach(k => {
        childrenWrapper.appendChild(createJsonTreeNode(value[k], k, currentPath, false));
      });

      const footer = document.createElement('div');
      footer.className = 'json-row json-object-footer';
      footer.style.display = autoExpand ? 'flex' : 'none';
      footer.innerHTML = `<span class="json-bracket">${closeBracket}</span>`;

      let expanded = autoExpand;
      chevron.addEventListener('click', (e) => {
        e.stopPropagation();
        expanded = !expanded;
        chevron.classList.toggle('collapsed', !expanded);
        childrenWrapper.style.display = expanded ? 'block' : 'none';
        footer.style.display = expanded ? 'flex' : 'none';
      });

      // Quick copy object json
      row.querySelector('[data-copy-json]').addEventListener('click', (e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(JSON.stringify(value, null, 2));
        const btn = e.currentTarget;
        btn.style.color = 'var(--success)';
        setTimeout(() => { btn.style.color = ''; }, 1200);
      });

      // Copy path
      row.querySelector('[data-copy-path]').addEventListener('click', (e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(currentPath);
        const btn = e.currentTarget;
        btn.style.color = 'var(--success)';
        setTimeout(() => { btn.style.color = ''; }, 1200);
      });

      nodeWrapper.appendChild(row);
      nodeWrapper.appendChild(childrenWrapper);
      nodeWrapper.appendChild(footer);
    } else {
      const row = document.createElement('div');
      row.className = 'json-row';

      let valHtml = '';
      if (type === 'string') {
        valHtml = `<span class="json-string">"${renderSafeHighlightedHtml(value)}"</span>`;
      } else if (type === 'number') {
        valHtml = `<span class="json-number">${value}</span>`;
      } else if (type === 'boolean') {
        valHtml = `<span class="json-boolean">${value}</span>`;
      }

      row.innerHTML = `
        <span class="json-key">${escapeHtml(key)}</span><span class="json-colon">:</span>
        ${valHtml}
        <div class="json-row-actions">
          <button class="btn-mini-copy" data-copy="${escapeHtml(String(value))}" title="Copy value">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
          </button>
          <button class="btn-mini-copy" data-copy-path="${escapeHtml(currentPath)}" title="Copy JSON path">
            <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
          </button>
        </div>
      `;
      attachMiniCopy(row);
      nodeWrapper.appendChild(row);
    }

    return nodeWrapper;
  }

  function attachMiniCopy(row) {
    row.querySelectorAll('.btn-mini-copy').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const copyVal = btn.getAttribute('data-copy');
        const copyPath = btn.getAttribute('data-copy-path');
        const text = copyVal !== null ? copyVal : copyPath;
        if (text !== null) {
          navigator.clipboard.writeText(text);
          btn.style.color = 'var(--success)';
          setTimeout(() => { btn.style.color = ''; }, 1200);
        }
      });
    });
  }

  // --- Table View Generator with Document CRUD ---
  function renderHitsTable(hits, columns) {
    if (!hits || hits.length === 0 || !columns || columns.length === 0) {
      el.tableContainer.innerHTML = '<div class="empty-state">No search hits found in response.</div>';
      return;
    }

    let html = '<table class="data-table"><thead><tr>';
    html += '<th style="width: 85px;">Action</th>';
    columns.forEach(col => {
      html += `<th>${escapeHtml(col)}</th>`;
    });
    html += '</tr></thead><tbody>';

    hits.forEach((row, idx) => {
      html += `<tr data-row-idx="${idx}">`;
      html += `<td><button class="btn btn-sm btn-secondary btn-inspect-row" data-row-idx="${idx}">
        <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
        <span>Inspect</span>
      </button></td>`;
      columns.forEach(col => {
        const val = row[col];
        let displayVal = '-';
        if (val !== undefined && val !== null) {
          if (typeof val === 'object') {
            displayVal = JSON.stringify(val);
          } else {
            displayVal = String(val);
          }
        }
        html += `<td title="${escapeHtml(displayVal)}">${renderSafeHighlightedHtml(displayVal)}</td>`;
      });
      html += '</tr>';
    });
    html += '</tbody></table>';

    el.tableContainer.innerHTML = html;

    // Attach inspect handlers
    el.tableContainer.querySelectorAll('.btn-inspect-row').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const rowIdx = parseInt(btn.getAttribute('data-row-idx'), 10);
        openDocInspector(hits[rowIdx]);
      });
    });
  }

  // --- Document Inspector & CRUD ---

  function openDocInspector(doc) {
    if (!doc) return;
    state.currentInspectDoc = doc;
    const docId = doc._id || '';
    const indexName = state.activeIdx || 'unknown';

    el.docInspectIndex.textContent = indexName;
    el.docInspectId.textContent = docId;

    // Clone doc source without metadata keys for editing
    const cleanDoc = {};
    for (const k in doc) {
      if (!k.startsWith('_')) {
        cleanDoc[k] = doc[k];
      }
    }

    el.docInspectContent.value = JSON.stringify(cleanDoc, null, 2);

    if (state.safeMode) {
      el.docInspectSafeNotice.classList.remove('hidden');
      el.btnSaveDoc.disabled = true;
      el.btnDeleteDoc.disabled = true;
      el.btnSaveDoc.title = 'Safe Mode is ON. Turn off Safe Mode in header to modify documents.';
      el.btnDeleteDoc.title = 'Safe Mode is ON. Turn off Safe Mode in header to delete documents.';
    } else {
      el.docInspectSafeNotice.classList.add('hidden');
      el.btnSaveDoc.disabled = false;
      el.btnDeleteDoc.disabled = false;
      el.btnSaveDoc.title = 'Save modifications';
      el.btnDeleteDoc.title = 'Delete this document permanently';
    }

    el.docInspectorModal.classList.add('active');
  }

  async function saveDocChanges() {
    if (state.safeMode) {
      alert('Safe Mode is ON. Cannot update document.');
      return;
    }
    const indexName = state.activeIdx;
    const docId = el.docInspectId.textContent;
    if (!indexName || !docId) {
      alert('Cannot identify target index and document ID.');
      return;
    }

    let parsed = {};
    try {
      parsed = JSON.parse(el.docInspectContent.value);
    } catch (e) {
      alert('Invalid JSON content: ' + e.message);
      return;
    }

    try {
      await apiPost('/api/docs/update', {
        index: indexName,
        id: docId,
        doc: parsed
      });
      alert('✓ Document updated successfully!');
      el.docInspectorModal.classList.remove('active');
      runCurrentQuery();
    } catch (err) {
      alert('Update failed: ' + err.message);
    }
  }

  async function deleteCurrentDoc() {
    if (state.safeMode) {
      alert('Safe Mode is ON. Cannot delete document.');
      return;
    }
    const indexName = state.activeIdx;
    const docId = el.docInspectId.textContent;
    if (!indexName || !docId) return;

    if (!confirm(`Are you sure you want to permanently delete document "${docId}" from index "${indexName}"?`)) {
      return;
    }

    try {
      await apiDelete(`/api/docs/${encodeURIComponent(indexName)}/${encodeURIComponent(docId)}`);
      alert(`✓ Document "${docId}" deleted successfully!`);
      el.docInspectorModal.classList.remove('active');
      runCurrentQuery();
    } catch (err) {
      alert('Delete failed: ' + err.message);
    }
  }

  // --- Visual Aggregations Chart Renderer ---

  function renderAggregationsCharts(aggregations) {
    el.aggChartContainer.innerHTML = '';

    for (const aggName in aggregations) {
      const aggData = aggregations[aggName];
      if (!aggData || !aggData.buckets) continue;

      const buckets = aggData.buckets;
      if (!Array.isArray(buckets) || buckets.length === 0) continue;

      const chartCard = document.createElement('div');
      chartCard.className = 'chart-card';

      let maxCount = 1;
      let totalCount = 0;
      buckets.forEach(b => {
        if (b.doc_count > maxCount) maxCount = b.doc_count;
        totalCount += b.doc_count;
      });

      let rowsHtml = '';
      buckets.forEach(b => {
        const key = b.key_as_string || b.key;
        const count = b.doc_count || 0;
        const pct = totalCount > 0 ? ((count / totalCount) * 100).toFixed(1) : 0;
        const widthPct = ((count / maxCount) * 100).toFixed(1);

        rowsHtml += `
          <div class="bar-chart-row" title="Click to filter by ${escapeHtml(key)}" data-filter-key="${escapeHtml(key)}">
            <div class="bar-label">${escapeHtml(key)}</div>
            <div class="bar-track">
              <div class="bar-fill" style="width: ${widthPct}%;"></div>
            </div>
            <div class="bar-count">${formatNumber(count)} (${pct}%)</div>
          </div>
        `;
      });

      chartCard.innerHTML = `
        <div class="chart-title">
          <span>Aggregation: <strong>${escapeHtml(aggName)}</strong></span>
          <span style="font-size:11px; color:var(--text-muted)">Total: ${formatNumber(totalCount)} docs</span>
        </div>
        <div class="chart-body">${rowsHtml}</div>
      `;

      el.aggChartContainer.appendChild(chartCard);
    }
  }

  // --- History & Snippets ---

  async function loadHistory() {
    try {
      const data = await apiGet('/api/history?limit=50');
      const items = data.history || [];
      if (items.length === 0) {
        el.historyList.innerHTML = '<div class="empty-state">No query history yet</div>';
        return;
      }
      el.historyList.innerHTML = '';
      items.forEach(h => {
        const card = document.createElement('div');
        card.className = 'history-card';
        card.innerHTML = `
          <div style="display:flex; justify-content:space-between; font-weight:600; font-size:11px; margin-bottom:2px;">
            <span><span class="tag-badge">${h.method || 'POST'}</span> ${escapeHtml(h.path || '')}</span>
            <span style="color:${h.status < 300 ? 'var(--success)' : 'var(--danger)'}">${h.status}</span>
          </div>
          <div style="font-size:10px; color:var(--text-muted); display:flex; justify-content:space-between;">
            <span>Took: ${h.took_ms}ms</span>
            <span>${new Date(h.timestamp).toLocaleTimeString()}</span>
          </div>
        `;
        card.addEventListener('click', () => {
          state.editorInstance.setValue(h.raw_input || '');
          if (h.index) setActiveIndex(h.index);
        });
        el.historyList.appendChild(card);
      });
    } catch (e) {
      console.warn('Failed to load history:', e);
    }
  }

  async function loadSnippets() {
    try {
      const data = await apiGet('/api/snippets');
      const snippets = data.snippets || [];
      renderSnippetsList(snippets);
    } catch (e) {
      console.warn('Failed to load snippets:', e);
    }
  }

  function renderSnippetsList(snippets) {
    if (snippets.length === 0) {
      el.snippetsList.innerHTML = '<div class="empty-state">No saved snippets</div>';
      return;
    }
    el.snippetsList.innerHTML = '';
    snippets.forEach(s => {
      const card = document.createElement('div');
      card.className = 'snippet-card';
      const tagsHtml = (s.tags || []).map(t => `<span class="tag-badge">#${escapeHtml(t)}</span>`).join('');
      card.innerHTML = `
        <div class="snippet-card-title">${escapeHtml(s.title)}</div>
        <div class="snippet-card-desc">${escapeHtml(s.description || '')}</div>
        <div class="card-tags">${tagsHtml}</div>
      `;
      card.addEventListener('click', () => {
        const text = s.content ? `${s.method} ${s.path}\n${s.content}` : `${s.method} ${s.path}`;
        state.editorInstance.setValue(text);
      });
      el.snippetsList.appendChild(card);
    });
  }

  async function loadDSLSnippets() {
    try {
      const data = await apiGet('/api/snippets/dsl');
      state.dslSnippets = data.snippets || [];
    } catch (e) {
      console.warn('Failed to load DSL snippets:', e);
    }
  }

  // --- Antigravity (Local CLI Bridge) ---

  function getEffectiveTargetIndex() {
    if (state.activeIdx) return state.activeIdx;
    if (state.editorInstance) {
      const firstLine = state.editorInstance.getValue().split('\n')[0].trim();
      const match = firstLine.match(/^(?:GET|POST|PUT|DELETE)\s+\/?([a-zA-Z0-9_\-\*]+)\/_search/i);
      if (match && match[1] && !match[1].startsWith('_')) return match[1];
    }
    return '';
  }

  async function checkAntigravityStatus() {
    try {
      const res = await apiGet('/api/antigravity/status');
      if (res && res.available) {
        if (el.antigravityDot) el.antigravityDot.style.background = 'var(--success)';
        if (el.btnAntigravity) el.btnAntigravity.title = `Antigravity (Sẵn sàng: ${res.version || 'Local CLI'}) - ⌘K`;
        if (el.agModalStatusBadge) {
          el.agModalStatusBadge.className = 'tag-badge status-green';
          el.agModalStatusBadge.textContent = '🟢 Đang hoạt động';
        }
        if (el.agModalPath) {
          el.agModalPath.textContent = `${res.cli_path} (${res.version || 'v2'})`;
        }
      } else {
        if (el.antigravityDot) el.antigravityDot.style.background = 'var(--danger)';
        if (el.btnAntigravity) el.btnAntigravity.title = 'Antigravity Local Bridge chưa sẵn sàng (Nhấn để xem hướng dẫn)';
        if (el.agModalStatusBadge) {
          el.agModalStatusBadge.className = 'tag-badge status-red';
          el.agModalStatusBadge.textContent = '🔴 Chưa phát hiện binary agy';
        }
        if (el.agModalPath && res && res.error) {
          el.agModalPath.textContent = res.error;
        }
      }
    } catch (e) {
      console.warn('Failed to check Antigravity status:', e);
      if (el.antigravityDot) el.antigravityDot.style.background = 'var(--danger)';
    }
  }

  function toggleAIPromptBar(forceOpen) {
    if (!el.aiPromptBar) return;
    const shouldOpen = forceOpen !== undefined ? forceOpen : el.aiPromptBar.classList.contains('hidden');
    if (shouldOpen) {
      el.aiPromptBar.classList.remove('hidden');
      if (el.aiClarificationBox) el.aiClarificationBox.classList.add('hidden');
      if (el.aiVerifiedBox) el.aiVerifiedBox.classList.add('hidden');

      // Update Schema Target badge on the prompt bar
      const targetIdx = getEffectiveTargetIndex();
      if (el.aiActiveIndexBadge) {
        if (targetIdx) {
          el.aiActiveIndexBadge.textContent = `Schema: ${targetIdx}`;
          el.aiActiveIndexBadge.style.color = 'var(--success)';
          el.aiActiveIndexBadge.title = `Antigravity sẽ sử dụng đúng schema mapping của index '${targetIdx}'`;
        } else {
          el.aiActiveIndexBadge.textContent = state.currentLang === 'vi' ? 'Schema: Chưa chọn index' : 'Schema: No index selected';
          el.aiActiveIndexBadge.style.color = 'var(--warning)';
          el.aiActiveIndexBadge.title = state.currentLang === 'vi' ? 'Khuyên dùng: Hãy chọn Target Index ở thanh tab để câu query sinh ra bám sát 100% schema' : 'Tip: Select Target Index to strictly align with schema mapping';
        }
      }

      if (el.aiPromptInput) {
        el.aiPromptInput.focus();
        el.aiPromptInput.select();
      }
    } else {
      el.aiPromptBar.classList.add('hidden');
    }
  }

  function setAILoading(isLoading, msg) {
    if (!el.aiLoadingIndicator) return;
    el.aiLoadingIndicator.classList.toggle('hidden', !isLoading);
    if (el.aiLoadingText && msg) {
      el.aiLoadingText.textContent = msg;
    }
    if (el.btnAIGenerate) {
      el.btnAIGenerate.disabled = isLoading;
    }
  }

  async function generateQueryWithAI(customPrompt) {
    const promptText = (customPrompt || (el.aiPromptInput ? el.aiPromptInput.value : '')).trim();
    if (!promptText) return;

    if (el.aiClarificationBox) el.aiClarificationBox.classList.add('hidden');
    if (el.aiVerifiedBox) el.aiVerifiedBox.classList.add('hidden');

    const targetIdx = getEffectiveTargetIndex();
    const dict = I18N[state.currentLang] || I18N.vi;
    const loadingMsg = targetIdx
      ? (state.currentLang === 'vi' ? `Antigravity đang phân tích schema của '${targetIdx}' và sinh Query DSL...` : `Antigravity is analyzing schema of '${targetIdx}' and generating Query DSL...`)
      : (dict.ai_loading_generate || 'Antigravity đang kiểm tra schema...');
    setAILoading(true, loadingMsg);

    try {
      const res = await apiPost('/api/antigravity/generate', {
        prompt: promptText,
        index: targetIdx,
        index_name: targetIdx,
        existing_query: state.editorInstance ? state.editorInstance.getValue() : ''
      });

      // Case 1: Schema requires clarification (ambiguous fields or no index selected)
      if (res && res.needs_clarification) {
        if (el.aiClarificationBox) el.aiClarificationBox.classList.remove('hidden');
        if (el.aiClarificationQuestion) {
          el.aiClarificationQuestion.textContent = res.question || (state.currentLang === 'vi' ? 'Cần làm rõ trường dữ liệu trong schema:' : 'Schema clarification needed:');
        }
        if (el.aiClarificationOptions) {
          el.aiClarificationOptions.innerHTML = '';
          const options = res.suggested_options || [];
          options.forEach(opt => {
            const btn = document.createElement('button');
            btn.type = 'button';
            btn.className = 'btn btn-xs btn-secondary';
            btn.style.cssText = 'border-color: #f59e0b; color: var(--text); background: rgba(245, 158, 11, 0.15); font-weight: 500;';
            btn.innerHTML = `<span style="color:#f59e0b; margin-right:4px;">✦</span> ${escapeHtml(opt.label || opt.field)}`;
            btn.addEventListener('click', () => {
              // If it was an index choice (when no index was selected)
              if (!targetIdx && opt.field) {
                setTargetIndex(opt.field, true);
                generateQueryWithAI(promptText);
                return;
              }

              // If it was a field replacement
              if (res.invalid_fields && res.invalid_fields.length > 0) {
                const invField = res.invalid_fields[0];
                let newPrompt = promptText;
                const regex = new RegExp(`\\b${invField}\\b`, 'gi');
                if (regex.test(newPrompt)) {
                  newPrompt = newPrompt.replace(regex, opt.field);
                } else {
                  newPrompt += ` (dùng trường '${opt.field}')`;
                }
                if (el.aiPromptInput) el.aiPromptInput.value = newPrompt;
                generateQueryWithAI(newPrompt);
              } else {
                const newPrompt = promptText + ` (sử dụng trường '${opt.field}')`;
                if (el.aiPromptInput) el.aiPromptInput.value = newPrompt;
                generateQueryWithAI(newPrompt);
              }
            });
            el.aiClarificationOptions.appendChild(btn);
          });
        }
        return;
      }

      // Case 2: Query DSL generated successfully and passed 100% schema validation
      if (res && res.query_dsl) {
        let newContent = res.query_dsl.trim();
        const currentEffectiveIdx = getEffectiveTargetIndex() || res.path?.replace(/^\/|\/_search$/g, '');
        // If the query doesn't have a REST method header and targetIdx is resolved, prepend it
        if (!/^(GET|POST|PUT|DELETE)\s+/i.test(newContent) && currentEffectiveIdx) {
          newContent = `POST /${currentEffectiveIdx}/_search\n${newContent}`;
        }
        if (state.editorInstance) {
          state.editorInstance.setValue(newContent);
          formatQueryText();
        }

        // Show verified schema notification
        if (el.aiVerifiedBox) {
          el.aiVerifiedBox.classList.remove('hidden');
          if (el.aiVerifiedText) {
            const verifiedList = (res.verified_fields && res.verified_fields.length > 0)
              ? ` [${res.verified_fields.join(', ')}]`
              : '';
            el.aiVerifiedText.textContent = state.currentLang === 'vi'
              ? `✓ Đã xác thực 100% schema trường truy vấn${verifiedList}`
              : `✓ 100% schema verified fields${verifiedList}`;
          }
        }

        // Auto-close prompt bar smoothly after a short pause
        setTimeout(() => {
          toggleAIPromptBar(false);
        }, 1800);
      } else {
        alert('Antigravity không trả về Query DSL hợp lệ.');
      }
    } catch (err) {
      alert('Lỗi sinh query với Antigravity: ' + err.message);
    } finally {
      setAILoading(false);
    }
  }

  async function explainCurrentQuery() {
    if (!state.editorInstance) return;
    const query = state.editorInstance.getValue().trim();
    if (!query) {
      alert(state.currentLang === 'vi' ? 'Vui lòng nhập hoặc mở truy vấn trước khi yêu cầu giải thích.' : 'Please enter or open a query first.');
      return;
    }

    const targetIdx = getEffectiveTargetIndex();
    if (el.aiExplainModal) el.aiExplainModal.classList.add('active');
    if (el.aiExplainContent) {
      el.aiExplainContent.innerHTML = `
        <div style="display:flex; align-items:center; gap:8px; color:var(--text-muted); padding:16px;">
          <span class="spinner"></span>
          <span>${state.currentLang === 'vi' ? 'Antigravity đang phân tích cấu trúc truy vấn DSL...' : 'Antigravity is analyzing query DSL structure...'}</span>
        </div>
      `;
    }

    try {
      const res = await apiPost('/api/antigravity/explain', {
        query: query,
        index: targetIdx,
        index_name: targetIdx
      });

      if (res && res.explanation) {
        let formatted = escapeHtml(res.explanation)
          .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
          .replace(/`([^`]+)`/g, '<code class="code-badge">$1</code>')
          .replace(/\n\n/g, '<br/><br/>')
          .replace(/\n/g, '<br/>');
        if (el.aiExplainContent) el.aiExplainContent.innerHTML = formatted;
      } else {
        if (el.aiExplainContent) el.aiExplainContent.textContent = 'Không nhận được phân tích từ Antigravity.';
      }
    } catch (err) {
      if (el.aiExplainContent) {
        el.aiExplainContent.innerHTML = `<span style="color:var(--danger)">Lỗi: ${escapeHtml(err.message)}</span>`;
      }
    }
  }

  async function fixCurrentQueryWithError(esError) {
    if (!state.editorInstance) return;
    const query = state.editorInstance.getValue().trim();
    if (!query) return;

    const targetIdx = getEffectiveTargetIndex();
    const dict = I18N[state.currentLang] || I18N.vi;
    toggleAIPromptBar(true);
    setAILoading(true, dict.ai_loading_fix || 'Antigravity đang chẩn đoán lỗi và sửa truy vấn...');

    try {
      const res = await apiPost('/api/antigravity/fix', {
        query: query,
        error: esError || '',
        es_error: esError || '',
        index: targetIdx,
        index_name: targetIdx
      });

      if (res && res.fixed_query) {
        let newContent = res.fixed_query.trim();
        if (!/^(GET|POST|PUT|DELETE)\s+/i.test(newContent) && targetIdx) {
          newContent = `POST /${targetIdx}/_search\n${newContent}`;
        }
        state.editorInstance.setValue(newContent);
        formatQueryText();
        toggleAIPromptBar(false);
        // Automatically rerun fixed query
        runCurrentQuery();
      } else {
        alert('Antigravity không thể tự động sửa lỗi truy vấn này.');
      }
    } catch (err) {
      alert('Lỗi tự động sửa truy vấn: ' + err.message);
    } finally {
      setAILoading(false);
    }
  }

  // --- Golang DSL Bidirectional Converter ---

  const GOLANG_DSL_STRUCTS_DEF = `type EsFiltersRange struct {
	Key string
	Min interface{}
	Max interface{}
}

type EsFilters struct {
	Type  string // support for term, wildcard and match
	Key   string
	Value any
}

const (
	ASC  = "asc"
	DESC = "desc"
)

type EsSort struct {
	Key   string
	Order string
}

type ESQuery map[string]interface{}
type ESQueryArr []map[string]interface{}
type EsFiltersArr []EsFilters
type EsRangesArr []EsFiltersRange
type EsSortArr []EsSort

func (f EsFiltersArr) Add(input EsFilters) EsFiltersArr {
	f = append(f, input)
	return f
}

func (f EsFiltersArr) Exists(field string) EsFiltersArr {
	f = append(f, EsFilters{Key: "field", Type: "exists", Value: field})
	return f
}

func (f EsFiltersArr) Query() ESQueryArr {
	var rs []map[string]interface{}
	for i := 0; i < len(f); i++ {
		if f[i].Type == "wildcard" {
			value := map[string]interface{}{}
			value["value"] = f[i].Value
			value["case_insensitive"] = true
			rs = append(rs, map[string]interface{}{
				f[i].Type: map[string]interface{}{
					f[i].Key: value,
				},
			})
		} else {
			rs = append(rs, map[string]interface{}{
				f[i].Type: map[string]interface{}{
					f[i].Key: f[i].Value,
				},
			})
		}
	}
	return rs
}

func (f EsRangesArr) Add(input EsFiltersRange) EsRangesArr {
	f = append(f, input)
	return f
}

func (f EsRangesArr) Query() ESQueryArr {
	var rs []map[string]interface{}
	for i := 0; i < len(f); i++ {
		rs = append(rs, map[string]interface{}{
			"range": map[string]interface{}{
				f[i].Key: map[string]interface{}{
					"gte": f[i].Min,
					"lte": f[i].Max,
				},
			},
		})
	}
	return rs
}

func (f EsSortArr) Add(input EsSort) EsSortArr {
	f = append(f, input)
	return f
}

func (f EsSortArr) Query() ESQueryArr {
	var rs []map[string]interface{}
	for i := 0; i < len(f); i++ {
		rs = append(rs, map[string]interface{}{
			f[i].Key: map[string]interface{}{
				"order": f[i].Order,
			},
		})
	}
	return rs
}

func (f ESQueryArr) Must() ESQuery {
	return ESQuery(map[string]interface{}{"must": f})
}

func (f ESQueryArr) MustNot() ESQuery {
	return ESQuery(map[string]interface{}{"must_not": f})
}

func (f ESQueryArr) Should() ESQuery {
	return ESQuery(map[string]interface{}{"should": f})
}

func (f ESQuery) Bool() ESQuery {
	return ESQuery(map[string]interface{}{"bool": f})
}

func (f ESQuery) Query() ESQuery {
	return ESQuery(map[string]interface{}{"query": f})
}

func (f ESQuery) Nested(nestedName string) ESQuery {
	return ESQuery(map[string]interface{}{
		"nested": map[string]interface{}{
			"path":  nestedName,
			"query": f,
		},
	})
}

func (f ESQuery) Sort(sort EsSortArr) ESQuery {
	f["sort"] = sort
	return f
}

func (f ESQuery) Paging(page, size int) ESQuery {
	f["size"] = size
	f["from"] = page * size
	return f
}

func QueryMatchAll() ESQuery {
	return ESQuery(map[string]interface{}{
		"match_all": map[string]interface{}{},
	}).Query()
}`;

  async function openGolangConverterModal() {
    if (!el.golangModal) return;
    el.golangModal.classList.add('active');
    switchConverterTab('queryToGo');

    const queryContent = state.editorInstance ? state.editorInstance.getValue().trim() : '';
    if (el.golangGeneratedOutput) {
      el.golangGeneratedOutput.textContent = '// Đang phân tích Query DSL và sinh mã nguồn Golang...';
    }

    try {
      const res = await apiPost('/api/converter/to-golang', {
        query_dsl: queryContent
      });
      if (res && res.golang_code) {
        if (el.golangGeneratedOutput) el.golangGeneratedOutput.textContent = res.golang_code;
      }
    } catch (err) {
      if (el.golangGeneratedOutput) {
        el.golangGeneratedOutput.textContent = `// Lỗi sinh mã nguồn: ${err.message}`;
      }
    }
  }

  function switchConverterTab(mode) {
    if (mode === 'queryToGo') {
      if (el.tabModeQueryToGo) {
        el.tabModeQueryToGo.className = 'btn btn-sm btn-primary';
      }
      if (el.tabModeGoToQuery) {
        el.tabModeGoToQuery.className = 'btn btn-sm btn-secondary';
      }
      if (el.panelQueryToGo) el.panelQueryToGo.classList.remove('hidden');
      if (el.panelGoToQuery) el.panelGoToQuery.classList.add('hidden');
    } else {
      if (el.tabModeQueryToGo) {
        el.tabModeQueryToGo.className = 'btn btn-sm btn-secondary';
      }
      if (el.tabModeGoToQuery) {
        el.tabModeGoToQuery.className = 'btn btn-sm btn-primary';
      }
      if (el.panelQueryToGo) el.panelQueryToGo.classList.add('hidden');
      if (el.panelGoToQuery) el.panelGoToQuery.classList.remove('hidden');
      if (el.golangInputCode) el.golangInputCode.focus();
    }
  }

  async function convertGoCodeToDSL() {
    if (!el.golangInputCode) return;
    const code = el.golangInputCode.value.trim();
    if (!code) {
      alert(state.currentLang === 'vi' ? 'Vui lòng nhập mã nguồn Golang' : 'Please input Golang code');
      return;
    }

    if (el.btnConvertGoToDSL) {
      el.btnConvertGoToDSL.disabled = true;
      el.btnConvertGoToDSL.innerHTML = '<span class="spinner"></span> Đang biên dịch...';
    }

    try {
      const res = await apiPost('/api/converter/to-dsl', {
        golang_code: code
      });
      if (res && res.query_dsl) {
        if (el.dslGeneratedOutput) el.dslGeneratedOutput.textContent = res.query_dsl;
        if (el.dslOutputWrapper) el.dslOutputWrapper.classList.remove('hidden');
        if (el.btnApplyDSLToEditor) el.btnApplyDSLToEditor.classList.remove('hidden');
      }
    } catch (err) {
      alert('Lỗi biên dịch: ' + err.message);
    } finally {
      if (el.btnConvertGoToDSL) {
        el.btnConvertGoToDSL.disabled = false;
        el.btnConvertGoToDSL.innerHTML = `
          <svg class="icon icon-sm" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"></polyline><polyline points="8 6 2 12 8 18"></polyline></svg>
          <span>Biên dịch sang Query DSL</span>
        `;
      }
    }
  }

  function applyDSLToEditor() {
    if (!el.dslGeneratedOutput || !state.editorInstance) return;
    const dsl = el.dslGeneratedOutput.textContent.trim();
    if (!dsl) return;

    let targetIdx = getEffectiveTargetIndex();
    let newContent = dsl;
    if (!/^(GET|POST|PUT|DELETE)\s+/i.test(dsl) && targetIdx) {
      newContent = `POST /${targetIdx}/_search\n${dsl}`;
    }
    state.editorInstance.setValue(newContent);
    formatQueryText();

    if (el.golangModal) el.golangModal.classList.remove('active');
  }

  // --- UI Event Listeners ---

  function setupEventListeners() {
    setupGRPCEventListeners();

    // Cluster Selector Switch
    el.clusterSelect.addEventListener('change', async (e) => {
      const targetId = e.target.value;
      if (!targetId) return;
      try {
        await apiPost(`/api/connections/${targetId}/activate`, {});
        await loadClusters();
      } catch (err) {
        alert('Failed to activate cluster: ' + err.message);
      }
    });

    // Editor Target Index Bar Select
    if (el.editorIndexSelect) {
      el.editorIndexSelect.addEventListener('change', (e) => {
        setTargetIndex(e.target.value, true);
      });
    }

    // Sync to Line 1 Button
    if (el.btnInsertEndpoint) {
      el.btnInsertEndpoint.addEventListener('click', () => {
        insertEndpointToLine1();
      });
    }

    // Language Toggle (VI / EN)
    if (el.btnToggleLang) {
      el.btnToggleLang.addEventListener('click', toggleLanguage);
    }

    // Theme Toggle Button
    if (el.btnToggleTheme) {
      el.btnToggleTheme.addEventListener('click', toggleTheme);
    }

    // Safe Mode Toggle
    el.btnSafeMode.addEventListener('click', () => {
      state.safeMode = !state.safeMode;
      const label = document.getElementById('safeModeLabel');
      const dict = I18N[state.currentLang] || I18N.vi;
      if (state.safeMode) {
        el.btnSafeMode.className = 'btn btn-sm btn-safe-on';
        if (label) label.textContent = dict.safe_mode_on;
        el.btnSafeMode.title = dict.safe_mode_title_on;
      } else {
        el.btnSafeMode.className = 'btn btn-sm btn-safe-off';
        if (label) label.textContent = dict.safe_mode_off;
        el.btnSafeMode.title = dict.safe_mode_title_off;
      }
    });

    // Linter Badge Click
    el.linterBadge.addEventListener('click', showLinterModal);
    el.btnCloseLinterModal.addEventListener('click', () => el.linterModal.classList.remove('active'));

    // JSON Tree Filter & Expand/Collapse All
    const treeFilter = document.getElementById('jsonTreeFilterInput');
    if (treeFilter) {
      treeFilter.addEventListener('input', () => {
        const term = treeFilter.value.toLowerCase().trim();
        const allRows = el.jsonTreeContainer.querySelectorAll('.json-row');
        allRows.forEach(r => {
          if (!term) {
            r.style.opacity = '1';
            r.style.backgroundColor = '';
          } else if (r.textContent.toLowerCase().includes(term)) {
            r.style.opacity = '1';
            r.style.backgroundColor = 'rgba(56, 189, 248, 0.15)';
          } else {
            r.style.opacity = '0.35';
            r.style.backgroundColor = '';
          }
        });
      });
    }

    // JSON Tree Scope Buttons
    if (el.btnScopeHits) {
      el.btnScopeHits.addEventListener('click', () => {
        state.jsonScope = 'hits';
        el.btnScopeHits.classList.add('active');
        if (el.btnScopeFull) el.btnScopeFull.classList.remove('active');
        renderJsonTree();
      });
    }

    if (el.btnScopeFull) {
      el.btnScopeFull.addEventListener('click', () => {
        state.jsonScope = 'full';
        el.btnScopeFull.classList.add('active');
        if (el.btnScopeHits) el.btnScopeHits.classList.remove('active');
        renderJsonTree();
      });
    }

    const btnExpandAll = document.getElementById('btnExpandAll');
    if (btnExpandAll) {
      btnExpandAll.addEventListener('click', () => {
        el.jsonTreeContainer.querySelectorAll('.json-toggle-chevron').forEach(c => c.classList.remove('collapsed'));
        el.jsonTreeContainer.querySelectorAll('.json-children-wrapper').forEach(w => w.style.display = 'block');
        el.jsonTreeContainer.querySelectorAll('.json-object-footer').forEach(f => f.style.display = 'flex');
      });
    }

    const btnCollapseAll = document.getElementById('btnCollapseAll');
    if (btnCollapseAll) {
      btnCollapseAll.addEventListener('click', () => {
        el.jsonTreeContainer.querySelectorAll('.json-toggle-chevron').forEach(c => c.classList.add('collapsed'));
        el.jsonTreeContainer.querySelectorAll('.json-children-wrapper').forEach(w => w.style.display = 'none');
        el.jsonTreeContainer.querySelectorAll('.json-object-footer').forEach(f => f.style.display = 'none');
      });
    }

    // Document Inspector Actions
    el.btnCloseDocInspector.addEventListener('click', () => el.docInspectorModal.classList.remove('active'));
    el.btnSaveDoc.addEventListener('click', saveDocChanges);
    el.btnDeleteDoc.addEventListener('click', deleteCurrentDoc);

    // Run Query Button
    el.btnRunQuery.addEventListener('click', runCurrentQuery);

    // Format Button
    el.btnFormat.addEventListener('click', formatQueryText);

    // Refresh Indices Button
    el.btnRefreshIndices.addEventListener('click', loadIndices);
    el.indicesSearchInput.addEventListener('input', () => renderIndices(state.indices));

    // Clear History Button
    el.btnClearHistory.addEventListener('click', async () => {
      if (confirm('Are you sure you want to clear query history?')) {
        await apiDelete('/api/history');
        await loadHistory();
      }
    });

    // Copy Result
    el.btnCopyResult.addEventListener('click', () => {
      if (!state.lastResult || !state.lastResult.raw_json) return;
      navigator.clipboard.writeText(state.lastResult.raw_json).then(() => {
        const span = el.btnCopyResult.querySelector('span');
        if (span) span.textContent = 'Copied';
        setTimeout(() => { if (span) span.textContent = 'Copy'; }, 2000);
      });
    });

    // Export CSV
    el.btnExportCSV.addEventListener('click', exportTableToCSV);

    // Tabs Management
    el.btnNewTab.addEventListener('click', () => {
      const newId = Date.now();
      const defaultContent = state.activeIdx
        ? `POST /${state.activeIdx}/_search\n{\n  "size": 20,\n  "query": {\n    "match_all": {}\n  }\n}`
        : '{\n  "size": 20,\n  "query": {\n    "match_all": {}\n  }\n}';
      state.tabs.push({
        id: newId,
        title: state.activeIdx ? `${state.activeIdx}` : `Query ${state.tabs.length + 1}`,
        index: state.activeIdx,
        content: defaultContent
      });
      renderEditorTabs();
      switchEditorTab(newId);
    });

    // Cluster Manager Modal
    el.btnManageClusters.addEventListener('click', openClusterModal);
    el.btnCloseClusterModal.addEventListener('click', () => el.clusterModal.classList.remove('active'));
    el.connAuthType.addEventListener('change', (e) => {
      const type = e.target.value;
      document.getElementById('authFieldBasic').classList.toggle('hidden', type !== 'basic');
      document.getElementById('authFieldAPIKey').classList.toggle('hidden', type !== 'apikey');
      document.getElementById('authFieldBearer').classList.toggle('hidden', type !== 'bearer');
    });

    el.btnTestConn.addEventListener('click', testClusterConnection);
    el.clusterForm.addEventListener('submit', saveClusterProfile);

    // Save Snippet Modal
    el.btnSaveSnippetModal.addEventListener('click', () => {
      el.saveSnippetModal.classList.add('active');
    });
    el.btnCloseSnippetModal.addEventListener('click', () => el.saveSnippetModal.classList.remove('active'));
    el.btnCancelSnippet.addEventListener('click', () => el.saveSnippetModal.classList.remove('active'));
    el.saveSnippetForm.addEventListener('submit', saveSnippetFromEditor);

    // Mapping Modal Close
    el.btnCloseMappingModal.addEventListener('click', () => el.mappingModal.classList.remove('active'));

    // Antigravity AI Prompt Bar Toggle & Shortcut (⌘K / Ctrl+K)
    window.addEventListener('keydown', (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        toggleAIPromptBar();
      }
      if (e.key === 'Escape') {
        if (el.aiPromptBar && !el.aiPromptBar.classList.contains('hidden')) {
          toggleAIPromptBar(false);
        }
      }
    });

    if (el.btnAntigravity) {
      el.btnAntigravity.addEventListener('click', (e) => {
        if (e.shiftKey || e.altKey) {
          if (el.antigravityModal) el.antigravityModal.classList.add('active');
        } else {
          toggleAIPromptBar();
        }
      });
    }

    if (el.btnCloseAIPromptBar) {
      el.btnCloseAIPromptBar.addEventListener('click', () => toggleAIPromptBar(false));
    }

    if (el.btnAIGenerate) {
      el.btnAIGenerate.addEventListener('click', () => generateQueryWithAI());
    }

    if (el.aiPromptInput) {
      el.aiPromptInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
          e.preventDefault();
          generateQueryWithAI();
        }
      });
    }

    // Quick prompt chips
    document.querySelectorAll('.ai-chip').forEach(chip => {
      chip.addEventListener('click', () => {
        const prompt = chip.getAttribute('data-prompt');
        if (el.aiPromptInput) el.aiPromptInput.value = prompt;
        generateQueryWithAI(prompt);
      });
    });

    // Explain query button
    if (el.btnAIExplain) {
      el.btnAIExplain.addEventListener('click', explainCurrentQuery);
    }

    if (el.btnCloseExplainModal) {
      el.btnCloseExplainModal.addEventListener('click', () => {
        if (el.aiExplainModal) el.aiExplainModal.classList.remove('active');
      });
    }

    // Antigravity Modal actions
    if (el.btnCloseAntigravityModal) {
      el.btnCloseAntigravityModal.addEventListener('click', () => {
        if (el.antigravityModal) el.antigravityModal.classList.remove('active');
      });
    }

    if (el.btnCheckAgyStatus) {
      el.btnCheckAgyStatus.addEventListener('click', async () => {
        el.btnCheckAgyStatus.disabled = true;
        el.btnCheckAgyStatus.textContent = 'Đang kiểm tra...';
        await checkAntigravityStatus();
        el.btnCheckAgyStatus.disabled = false;
        el.btnCheckAgyStatus.textContent = 'Kiểm tra kết nối';
      });
    }

    if (el.btnOpenAIPromptFromModal) {
      el.btnOpenAIPromptFromModal.addEventListener('click', () => {
        if (el.antigravityModal) el.antigravityModal.classList.remove('active');
        toggleAIPromptBar(true);
      });
    }

    // Golang Converter Modal Events
    if (el.btnGolangModal) {
      el.btnGolangModal.addEventListener('click', openGolangConverterModal);
    }
    if (el.btnCloseGolangModal) {
      el.btnCloseGolangModal.addEventListener('click', () => {
        if (el.golangModal) el.golangModal.classList.remove('active');
      });
    }
    if (el.tabModeQueryToGo) {
      el.tabModeQueryToGo.addEventListener('click', () => switchConverterTab('queryToGo'));
    }
    if (el.tabModeGoToQuery) {
      el.tabModeGoToQuery.addEventListener('click', () => switchConverterTab('goToQuery'));
    }
    if (el.btnCopyGolangCode) {
      el.btnCopyGolangCode.addEventListener('click', () => {
        if (!el.golangGeneratedOutput) return;
        const text = el.golangGeneratedOutput.textContent;
        navigator.clipboard.writeText(text).then(() => {
          const span = el.btnCopyGolangCode.querySelector('span');
          const originalText = span ? span.textContent : '';
          if (span) span.textContent = state.currentLang === 'vi' ? 'Đã sao chép ✓' : 'Copied ✓';
          setTimeout(() => { if (span) span.textContent = originalText; }, 2000);
        });
      });
    }
    if (el.btnCopyGolangStructs) {
      el.btnCopyGolangStructs.addEventListener('click', () => {
        navigator.clipboard.writeText(GOLANG_DSL_STRUCTS_DEF).then(() => {
          const span = el.btnCopyGolangStructs.querySelector('span');
          const originalText = span ? span.textContent : '';
          if (span) span.textContent = state.currentLang === 'vi' ? 'Đã sao chép Structs ✓' : 'Copied Structs ✓';
          setTimeout(() => { if (span) span.textContent = originalText; }, 2000);
        });
      });
    }
    if (el.btnConvertGoToDSL) {
      el.btnConvertGoToDSL.addEventListener('click', convertGoCodeToDSL);
    }
    if (el.btnApplyDSLToEditor) {
      el.btnApplyDSLToEditor.addEventListener('click', applyDSLToEditor);
    }
  }

  // --- Editor Tabs Handling ---

  function renderEditorTabs() {
    el.editorTabsList.innerHTML = '';
    state.tabs.forEach(tab => {
      const tabEl = document.createElement('div');
      tabEl.className = `editor-tab ${tab.id === state.activeTabId ? 'active' : ''}`;
      tabEl.innerHTML = `
        <span class="tab-title">${escapeHtml(tab.title)}</span>
        ${state.tabs.length > 1 ? '<span class="tab-close">&times;</span>' : ''}
      `;
      tabEl.addEventListener('click', (e) => {
        if (e.target.classList.contains('tab-close')) {
          closeEditorTab(tab.id);
        } else {
          switchEditorTab(tab.id);
        }
      });
      el.editorTabsList.appendChild(tabEl);
    });
  }

  function switchEditorTab(tabId) {
    state.activeTabId = tabId;
    const tab = state.tabs.find(t => t.id === tabId);
    if (tab && state.editorInstance) {
      state.editorInstance.setValue(tab.content);
      if (tab.index) {
        setTargetIndex(tab.index, false);
      } else {
        setTargetIndex('', false);
      }
    }
    renderEditorTabs();
  }

  function closeEditorTab(tabId) {
    if (state.tabs.length <= 1) return;
    const idx = state.tabs.findIndex(t => t.id === tabId);
    state.tabs = state.tabs.filter(t => t.id !== tabId);
    if (state.activeTabId === tabId) {
      const nextTab = state.tabs[Math.max(0, idx - 1)];
      switchEditorTab(nextTab.id);
    } else {
      renderEditorTabs();
    }
  }

  // --- Sidebar & View Tabs ---

  function setupSidebarTabs() {
    el.sidebarTabs.forEach(btn => {
      btn.addEventListener('click', () => {
        el.sidebarTabs.forEach(b => b.classList.remove('active'));
        el.sidebarContents.forEach(c => c.classList.remove('active'));
        btn.classList.add('active');
        const target = btn.getAttribute('data-tab');
        document.getElementById(target).classList.add('active');
      });
    });
  }

  function setupViewTabs() {
    el.viewTabs.forEach(btn => {
      btn.addEventListener('click', () => {
        const view = btn.getAttribute('data-view');
        switchResultView(view);
      });
    });
  }

  function switchResultView(view) {
    el.viewTabs.forEach(b => {
      b.classList.toggle('active', b.getAttribute('data-view') === view);
    });
    el.resultPanels.forEach(p => {
      p.classList.toggle('active', p.id === `view-${view}`);
    });
    state.currentView = view;
  }

  // --- Split Resizer ---
  function setupResizer() {
    let isResizing = false;
    el.paneResizer.addEventListener('mousedown', (e) => {
      isResizing = true;
      document.body.style.cursor = 'row-resize';
      e.preventDefault();
    });

    document.addEventListener('mousemove', (e) => {
      if (!isResizing) return;
      const containerHeight = el.paneResizer.parentElement.clientHeight;
      const offsetTop = el.paneResizer.parentElement.getBoundingClientRect().top;
      const newHeight = e.clientY - offsetTop;
      if (newHeight > 100 && newHeight < containerHeight - 100) {
        el.editorPane.style.height = `${newHeight}px`;
      }
    });

    document.addEventListener('mouseup', () => {
      if (isResizing) {
        isResizing = false;
        document.body.style.cursor = '';
      }
    });
  }

  // --- Cluster Modal Handlers ---

  function openClusterModal() {
    el.clusterModal.classList.add('active');
    renderClusterProfiles();
  }

  function renderClusterProfiles() {
    el.clusterProfilesList.innerHTML = '';
    state.clusters.forEach(c => {
      const item = document.createElement('div');
      item.className = `cluster-item ${c.id === (state.activeCluster ? state.activeCluster.id : '') ? 'active' : ''}`;
      item.innerHTML = `
        <div>
          <div class="cluster-item-title">${escapeHtml(c.name)} ${c.id === (state.activeCluster ? state.activeCluster.id : '') ? '<span class="tag-badge" style="color:var(--accent); margin-left:6px;">Active</span>' : ''}</div>
          <div class="cluster-item-url">${escapeHtml(c.url)} [Auth: ${c.auth_type}]</div>
        </div>
        <div style="display:flex; gap:6px;">
          <button class="btn btn-sm btn-secondary btn-edit-conn">Edit</button>
          <button class="btn btn-sm btn-danger btn-del-conn">Delete</button>
        </div>
      `;

      item.querySelector('.btn-edit-conn').addEventListener('click', () => {
        document.getElementById('connId').value = c.id;
        document.getElementById('connName').value = c.name;
        document.getElementById('connUrl').value = c.url;
        document.getElementById('connAuthType').value = c.auth_type;
        document.getElementById('connAuthType').dispatchEvent(new Event('change'));
        document.getElementById('connUsername').value = c.username || '';
        document.getElementById('connPassword').value = c.password || '';
        document.getElementById('connApiKey').value = c.api_key || '';
        document.getElementById('connBearer').value = c.bearer_token || '';
        document.getElementById('connInsecure').checked = c.insecure_skip_verify;
      });

      item.querySelector('.btn-del-conn').addEventListener('click', async () => {
        if (confirm(`Delete connection profile ${c.name}?`)) {
          await apiDelete(`/api/connections/${c.id}`);
          await loadClusters();
          renderClusterProfiles();
        }
      });

      el.clusterProfilesList.appendChild(item);
    });
  }

  async function testClusterConnection() {
    const payload = getClusterFormData();
    el.testConnResult.className = 'test-result-box';
    el.testConnResult.style.display = 'block';
    el.testConnResult.textContent = 'Testing connection...';

    try {
      const res = await apiPost('/api/connections/test', payload);
      if (res.success) {
        const v = res.info && res.info.version ? res.info.version.number : '';
        const h = res.health ? res.health.status : 'OK';
        el.testConnResult.className = 'test-result-box success';
        el.testConnResult.textContent = `✓ Connected successfully! Elasticsearch ${v}, Health: ${h}`;
      } else {
        el.testConnResult.className = 'test-result-box error';
        el.testConnResult.textContent = `✗ Connection failed: ${res.error}`;
      }
    } catch (err) {
      el.testConnResult.className = 'test-result-box error';
      el.testConnResult.textContent = `✗ Error: ${err.message}`;
    }
  }

  async function saveClusterProfile(e) {
    e.preventDefault();
    const payload = getClusterFormData();
    try {
      await apiPost('/api/connections', payload);
      el.clusterForm.reset();
      document.getElementById('connId').value = '';
      el.testConnResult.style.display = 'none';
      await loadClusters();
      renderClusterProfiles();
    } catch (err) {
      alert('Failed to save profile: ' + err.message);
    }
  }

  function getClusterFormData() {
    return {
      id: document.getElementById('connId').value,
      name: document.getElementById('connName').value,
      url: document.getElementById('connUrl').value,
      auth_type: document.getElementById('connAuthType').value,
      username: document.getElementById('connUsername').value,
      password: document.getElementById('connPassword').value,
      api_key: document.getElementById('connApiKey').value,
      bearer_token: document.getElementById('connBearer').value,
      insecure_skip_verify: document.getElementById('connInsecure').checked,
      timeout_seconds: 30
    };
  }

  // --- Snippet Save Handler ---
  async function saveSnippetFromEditor(e) {
    e.preventDefault();
    const title = document.getElementById('snipTitle').value;
    const desc = document.getElementById('snipDesc').value;
    const tags = document.getElementById('snipTags').value.split(',').map(t => t.trim()).filter(Boolean);
    const content = state.editorInstance.getValue();

    try {
      await apiPost('/api/snippets', {
        title: title,
        description: desc,
        tags: tags,
        method: 'POST',
        path: state.activeIdx ? `/${state.activeIdx}/_search` : '/_search',
        content: content
      });
      el.saveSnippetModal.classList.remove('active');
      el.saveSnippetForm.reset();
      await loadSnippets();
    } catch (err) {
      alert('Failed to save snippet: ' + err.message);
    }
  }

  // --- CSV Export Helper ---
  function exportTableToCSV() {
    if (!state.lastResult || !state.lastResult.extracted_hits || state.lastResult.extracted_hits.length === 0) {
      alert('No data hits to export');
      return;
    }

    const hits = state.lastResult.extracted_hits;
    const columns = state.lastResult.hit_columns;

    const csvRows = [];
    csvRows.push(columns.map(c => `"${c.replace(/"/g, '""')}"`).join(','));

    hits.forEach(row => {
      const rowVals = columns.map(col => {
        let val = row[col];
        if (val === undefined || val === null) val = '';
        else if (typeof val === 'object') val = JSON.stringify(val);
        else val = String(val);
        return `"${val.replace(/"/g, '""')}"`;
      });
      csvRows.push(rowVals.join(','));
    });

    const csvContent = 'data:text/csv;charset=utf-8,' + encodeURIComponent(csvRows.join('\n'));
    const link = document.createElement('a');
    link.setAttribute('href', csvContent);
    link.setAttribute('download', `eskhan_export_${Date.now()}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  }

  // --- Utility Helpers ---
  function validateJSON(text) {
    const trimmed = text.trim();
    if (!trimmed) {
      el.editorSyntaxStatus.textContent = state.currentLang === 'vi' ? 'Trống' : 'Empty';
      el.editorSyntaxStatus.style.color = 'var(--text-muted)';
      return;
    }

    let jsonText = trimmed;
    const lines = trimmed.split('\n');
    const firstLine = lines[0].trim();
    if (/^(GET|POST|PUT|DELETE)\s+/i.test(firstLine) || firstLine.startsWith('/')) {
      jsonText = lines.slice(1).join('\n').trim();
    }

    if (!jsonText) {
      el.editorSyntaxStatus.textContent = 'REST Endpoint';
      el.editorSyntaxStatus.style.color = 'var(--accent)';
      return;
    }

    try {
      JSON.parse(jsonText);
      el.editorSyntaxStatus.textContent = state.currentLang === 'vi' ? 'JSON Hợp lệ ✓' : 'JSON Valid ✓';
      el.editorSyntaxStatus.style.color = 'var(--success)';
    } catch (e) {
      el.editorSyntaxStatus.textContent = state.currentLang === 'vi' ? 'Lỗi cú pháp JSON ✗' : 'JSON Invalid ✗';
      el.editorSyntaxStatus.style.color = 'var(--danger)';
    }
  }

  function formatQueryText() {
    const content = state.editorInstance.getValue().trim();
    if (!content) return;

    const lines = content.split('\n');
    const firstLine = lines[0].trim();
    let header = '';
    let jsonPart = content;

    if (/^(GET|POST|PUT|DELETE)\s+/i.test(firstLine) || firstLine.startsWith('/')) {
      header = firstLine + '\n';
      jsonPart = lines.slice(1).join('\n').trim();
    }

    if (jsonPart) {
      try {
        const parsed = JSON.parse(jsonPart);
        const pretty = JSON.stringify(parsed, null, 2);
        state.editorInstance.setValue(header + pretty);
      } catch (e) {
        console.warn('Cannot format invalid JSON:', e);
      }
    }
  }

  function escapeHtml(str) {
    if (!str) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  function renderSafeHighlightedHtml(str) {
    if (!str) return '';
    const escaped = escapeHtml(str);
    // Convert escaped <mark> tags into styled high-contrast highlight spans
    return escaped.replace(/&lt;mark&gt;/g, '<mark class="es-hl">').replace(/&lt;\/mark&gt;/g, '</mark>');
  }

  function formatNumber(num) {
    if (num === undefined || num === null) return '0';
    return Number(num).toLocaleString();
  }

  function formatRawJSON(raw) {
    try {
      const parsed = JSON.parse(raw);
      return JSON.stringify(parsed, null, 2);
    } catch (e) {
      return raw;
    }
  }

  // ==========================================================================
  // v2: gRPC Studio Controller & Dynamic Invoker Logic
  // ==========================================================================

  const grpcState = {
    currentMode: 'es', // 'es' or 'grpc'
    services: [],
    selectedService: null,
    selectedMethod: null,
    editor: null,
    respEditor: null,
    isEditorInitialized: false
  };

  function switchStudioMode(mode) {
    grpcState.currentMode = mode;
    if (mode === 'es') {
      if (el.btnModeES) el.btnModeES.classList.add('active');
      if (el.btnModeGRPC) el.btnModeGRPC.classList.remove('active');
      if (el.esWorkspace) el.esWorkspace.classList.remove('hidden');
      if (el.grpcWorkspace) el.grpcWorkspace.classList.add('hidden');
      if (el.clusterSelectorContainer) el.clusterSelectorContainer.classList.remove('hidden');
      if (el.btnSafeMode) el.btnSafeMode.classList.remove('hidden');
      if (el.btnAntigravity) el.btnAntigravity.classList.remove('hidden');
      if (el.btnFormat) el.btnFormat.classList.remove('hidden');
      if (el.btnSaveSnippetModal) el.btnSaveSnippetModal.classList.remove('hidden');
      if (el.btnRunQuery) el.btnRunQuery.classList.remove('hidden');
      if (state.editorInstance) {
        setTimeout(() => state.editorInstance.layout(), 60);
      }
    } else if (mode === 'grpc') {
      if (el.btnModeGRPC) el.btnModeGRPC.classList.add('active');
      if (el.btnModeES) el.btnModeES.classList.remove('active');
      if (el.grpcWorkspace) el.grpcWorkspace.classList.remove('hidden');
      if (el.esWorkspace) el.esWorkspace.classList.add('hidden');
      if (el.clusterSelectorContainer) el.clusterSelectorContainer.classList.add('hidden');
      if (el.btnSafeMode) el.btnSafeMode.classList.add('hidden');
      if (el.btnAntigravity) el.btnAntigravity.classList.add('hidden');
      if (el.btnFormat) el.btnFormat.classList.add('hidden');
      if (el.btnSaveSnippetModal) el.btnSaveSnippetModal.classList.add('hidden');
      if (el.btnRunQuery) el.btnRunQuery.classList.add('hidden');

      initGRPCMonaco();
      setTimeout(() => {
        if (grpcState.editor) grpcState.editor.layout();
        if (grpcState.respEditor) grpcState.respEditor.layout();
      }, 60);
    }
  }

  function initGRPCMonaco() {
    if (grpcState.isEditorInitialized || !state.isMonacoLoaded || !window.monaco) return;

    const themeName = state.theme === 'light' ? 'vs' : 'vs-dark';

    // Request Monaco Editor
    if (el.grpcMonacoContainer) {
      grpcState.editor = monaco.editor.create(el.grpcMonacoContainer, {
        value: '{\n  \n}',
        language: 'json',
        theme: themeName,
        automaticLayout: true,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        fontSize: 13,
        tabSize: 2,
        renderLineHighlight: 'all',
        formatOnPaste: true,
        formatOnType: true
      });

      grpcState.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, function () {
        invokeGRPC();
      });

      grpcState.editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyF, function () {
        formatGRPCBody();
      });
    }

    // Response Monaco Editor (Read-Only)
    if (el.grpcRespMonacoContainer) {
      grpcState.respEditor = monaco.editor.create(el.grpcRespMonacoContainer, {
        value: '// Response will appear here after Invoke RPC',
        language: 'json',
        theme: themeName,
        automaticLayout: true,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        fontSize: 13,
        tabSize: 2,
        readOnly: true
      });
    }

    grpcState.isEditorInitialized = true;
  }

  function formatGRPCBody() {
    if (!grpcState.editor) return;
    try {
      const val = grpcState.editor.getValue();
      if (!val.trim()) return;
      const parsed = JSON.parse(val);
      grpcState.editor.setValue(JSON.stringify(parsed, null, 2));
    } catch (e) {
      showToast('Lỗi format JSON: ' + e.message, 'warning');
    }
  }

  function resetGRPCMock() {
    if (!grpcState.selectedMethod) {
      showToast('Chưa chọn Method để khôi phục Mock payload', 'warning');
      return;
    }
    if (grpcState.editor && grpcState.selectedMethod.mock_request) {
      grpcState.editor.setValue(grpcState.selectedMethod.mock_request);
      showToast('Đã khôi phục Mock payload mẫu', 'info');
    }
  }

  async function reflectGRPCServices() {
    const target = el.grpcTargetInput.value.trim();
    if (!target) {
      showToast('Vui lòng nhập địa chỉ gRPC target (ví dụ: localhost:50051)', 'warning');
      return;
    }

    const origBtnHtml = el.btnGrpcReflect.innerHTML;
    el.btnGrpcReflect.disabled = true;
    el.btnGrpcReflect.innerHTML = `<span class="spinner"></span> <span>Reflecting...</span>`;

    try {
      const res = await fetch('/api/grpc/reflect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          target: target,
          plaintext: el.grpcPlaintextCheck.checked,
          insecure_skip_verify: el.grpcInsecureCheck.checked
        })
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Server Reflection failed');
      }

      grpcState.services = data.services || [];
      populateGRPCServices(grpcState.services);
      showToast(`Đã khám phá thành công ${grpcState.services.length} services!`, 'success');
    } catch (err) {
      showToast('gRPC Reflection error: ' + err.message, 'danger');
    } finally {
      el.btnGrpcReflect.disabled = false;
      el.btnGrpcReflect.innerHTML = origBtnHtml;
    }
  }

  function populateGRPCServices(services) {
    if (!el.grpcServiceSelect) return;
    el.grpcServiceSelect.innerHTML = '';
    if (!services || services.length === 0) {
      el.grpcServiceSelect.innerHTML = '<option value="">(No services found)</option>';
      if (el.grpcMethodSelect) el.grpcMethodSelect.innerHTML = '<option value="">(No methods)</option>';
      return;
    }

    services.forEach(s => {
      const opt = document.createElement('option');
      opt.value = s.name;
      opt.textContent = s.name;
      el.grpcServiceSelect.appendChild(opt);
    });

    onGRPCServiceChange();
  }

  function onGRPCServiceChange() {
    const selectedSvcName = el.grpcServiceSelect.value;
    const svc = grpcState.services.find(s => s.name === selectedSvcName);
    grpcState.selectedService = svc || null;

    if (!el.grpcMethodSelect) return;
    el.grpcMethodSelect.innerHTML = '';
    if (!svc || !svc.methods || svc.methods.length === 0) {
      el.grpcMethodSelect.innerHTML = '<option value="">(No methods available)</option>';
      return;
    }

    svc.methods.forEach(m => {
      const opt = document.createElement('option');
      opt.value = m.name;
      opt.textContent = `${m.name} (${m.input_type.split('.').pop()} → ${m.output_type.split('.').pop()})`;
      el.grpcMethodSelect.appendChild(opt);
    });

    onGRPCMethodChange();
  }

  function onGRPCMethodChange() {
    if (!grpcState.selectedService || !el.grpcMethodSelect) return;
    const methodName = el.grpcMethodSelect.value;
    const m = (grpcState.selectedService.methods || []).find(method => method.name === methodName);
    grpcState.selectedMethod = m || null;

    if (m && m.mock_request && grpcState.editor) {
      grpcState.editor.setValue(m.mock_request);
    }
  }

  async function invokeGRPC() {
    const target = el.grpcTargetInput ? el.grpcTargetInput.value.trim() : '';
    const service = el.grpcServiceSelect ? el.grpcServiceSelect.value : '';
    const method = el.grpcMethodSelect ? el.grpcMethodSelect.value : '';

    if (!target) {
      showToast('Target host:port cannot be empty', 'warning');
      return;
    }
    if (!service || !method) {
      showToast('Please select a Service and Method first', 'warning');
      return;
    }

    let bodyStr = '{}';
    if (grpcState.editor) {
      bodyStr = grpcState.editor.getValue();
    }

    // Collect metadata
    const metadata = {};
    if (el.grpcMetaRows) {
      const rows = el.grpcMetaRows.querySelectorAll('tr');
      rows.forEach(tr => {
        const keyInput = tr.querySelector('.meta-key-input');
        const valInput = tr.querySelector('.meta-val-input');
        if (keyInput && valInput && keyInput.value.trim()) {
          metadata[keyInput.value.trim()] = valInput.value.trim();
        }
      });
    }

    const timeoutMs = parseInt(el.grpcTimeoutInput ? el.grpcTimeoutInput.value : 10000) || 10000;

    // UI updating
    if (el.grpcStatusBadge) el.grpcStatusBadge.className = 'grpc-status-badge';
    if (el.grpcStatusDot) el.grpcStatusDot.className = 'status-dot dot-yellow';
    if (el.grpcStatusText) el.grpcStatusText.textContent = 'Invoking...';
    if (el.grpcLatencyVal) el.grpcLatencyVal.textContent = '...';
    if (el.grpcSizeVal) el.grpcSizeVal.textContent = '...';
    if (el.btnGrpcInvoke) el.btnGrpcInvoke.disabled = true;

    try {
      const res = await fetch('/api/grpc/invoke', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          target: target,
          plaintext: el.grpcPlaintextCheck ? el.grpcPlaintextCheck.checked : true,
          insecure_skip_verify: el.grpcInsecureCheck ? el.grpcInsecureCheck.checked : false,
          service: service,
          method: method,
          body: bodyStr,
          metadata: metadata,
          timeout_ms: timeoutMs
        })
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Invoke request failed');
      }

      if (data.status_code === 'OK') {
        if (el.grpcStatusBadge) el.grpcStatusBadge.className = 'grpc-status-badge status-ok';
        if (el.grpcStatusDot) el.grpcStatusDot.className = 'status-dot dot-green';
        if (el.grpcStatusText) el.grpcStatusText.textContent = 'OK (0)';
      } else {
        if (el.grpcStatusBadge) el.grpcStatusBadge.className = 'grpc-status-badge status-error';
        if (el.grpcStatusDot) el.grpcStatusDot.className = 'status-dot dot-red';
        if (el.grpcStatusText) el.grpcStatusText.textContent = `${data.status_code} (${data.code}): ${data.message || ''}`;
      }

      if (el.grpcLatencyVal) el.grpcLatencyVal.textContent = `${data.took_ms} ms`;
      const sizeBytes = data.raw_json ? new Blob([data.raw_json]).size : 0;
      if (el.grpcSizeVal) {
        el.grpcSizeVal.textContent = sizeBytes > 1024 ? `${(sizeBytes / 1024).toFixed(1)} KB` : `${sizeBytes} B`;
      }

      if (grpcState.respEditor) {
        if (data.raw_json) {
          grpcState.respEditor.setValue(data.raw_json);
        } else if (data.message) {
          grpcState.respEditor.setValue(`// Status: ${data.status_code}\n// Code: ${data.code}\n// Message: ${data.message}`);
        } else {
          grpcState.respEditor.setValue(JSON.stringify(data, null, 2));
        }
      }

      renderGRPCHeaders(data.headers, data.trailers);
    } catch (err) {
      if (el.grpcStatusBadge) el.grpcStatusBadge.className = 'grpc-status-badge status-error';
      if (el.grpcStatusDot) el.grpcStatusDot.className = 'status-dot dot-red';
      if (el.grpcStatusText) el.grpcStatusText.textContent = 'Error: ' + err.message;
      if (grpcState.respEditor) {
        grpcState.respEditor.setValue(`// Error invoking RPC:\n${err.message}`);
      }
      showToast('Invoke error: ' + err.message, 'danger');
    } finally {
      if (el.btnGrpcInvoke) el.btnGrpcInvoke.disabled = false;
    }
  }

  function renderGRPCHeaders(headers, trailers) {
    if (!el.grpcRespHeadersList) return;
    el.grpcRespHeadersList.innerHTML = '';
    const hasHeaders = headers && Object.keys(headers).length > 0;
    const hasTrailers = trailers && Object.keys(trailers).length > 0;

    if (!hasHeaders && !hasTrailers) {
      el.grpcRespHeadersList.innerHTML = '<div class="empty-state">No headers or trailers returned</div>';
      return;
    }

    if (hasHeaders) {
      const headerTitle = document.createElement('div');
      headerTitle.style.fontWeight = 'bold';
      headerTitle.style.marginBottom = '6px';
      headerTitle.style.color = 'var(--text-main)';
      headerTitle.textContent = 'Response Headers:';
      el.grpcRespHeadersList.appendChild(headerTitle);

      for (const [k, v] of Object.entries(headers)) {
        const row = document.createElement('div');
        row.className = 'grpc-header-item';
        row.innerHTML = `<span class="grpc-header-key">${escapeHtml(k)}:</span> <span class="grpc-header-val">${escapeHtml(Array.isArray(v) ? v.join(', ') : v)}</span>`;
        el.grpcRespHeadersList.appendChild(row);
      }
    }

    if (hasTrailers) {
      const trailerTitle = document.createElement('div');
      trailerTitle.style.fontWeight = 'bold';
      trailerTitle.style.marginTop = '12px';
      trailerTitle.style.marginBottom = '6px';
      trailerTitle.style.color = 'var(--text-main)';
      trailerTitle.textContent = 'Response Trailers:';
      el.grpcRespHeadersList.appendChild(trailerTitle);

      for (const [k, v] of Object.entries(trailers)) {
        const row = document.createElement('div');
        row.className = 'grpc-header-item';
        row.innerHTML = `<span class="grpc-header-key">${escapeHtml(k)}:</span> <span class="grpc-header-val">${escapeHtml(Array.isArray(v) ? v.join(', ') : v)}</span>`;
        el.grpcRespHeadersList.appendChild(row);
      }
    }
  }

  function addGRPCMetaRow(key = '', val = '') {
    if (!el.grpcMetaRows) return;
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><input type="text" class="meta-key-input" placeholder="e.g. authorization" value="${escapeHtml(key)}"></td>
      <td><input type="text" class="meta-val-input" placeholder="e.g. Bearer token" value="${escapeHtml(val)}"></td>
      <td style="text-align:center;"><button type="button" class="btn btn-xs btn-danger btn-del-meta" title="Remove Header">&times;</button></td>
    `;
    tr.querySelector('.btn-del-meta').addEventListener('click', () => {
      tr.remove();
      updateGRPCMetaCount();
    });
    tr.querySelectorAll('input').forEach(inp => {
      inp.addEventListener('input', updateGRPCMetaCount);
    });
    el.grpcMetaRows.appendChild(tr);
    updateGRPCMetaCount();
  }

  function updateGRPCMetaCount() {
    if (!el.grpcMetaRows || !el.grpcMetaCount) return;
    let count = 0;
    el.grpcMetaRows.querySelectorAll('tr').forEach(tr => {
      const k = tr.querySelector('.meta-key-input');
      if (k && k.value.trim()) count++;
    });
    el.grpcMetaCount.textContent = count;
  }

  function copyGRPCResponse() {
    if (!grpcState.respEditor) return;
    const text = grpcState.respEditor.getValue();
    navigator.clipboard.writeText(text).then(() => {
      showToast('Đã copy Response JSON vào clipboard!', 'success');
    }).catch(err => {
      showToast('Lỗi copy: ' + err.message, 'danger');
    });
  }

  function openGrpcProtoModal() {
    if (el.grpcProtoModal) el.grpcProtoModal.classList.remove('hidden');
  }

  function closeGrpcProtoModal() {
    if (el.grpcProtoModal) el.grpcProtoModal.classList.add('hidden');
  }

  async function parseProtoDefinition() {
    const content = el.grpcProtoInput ? el.grpcProtoInput.value.trim() : '';
    if (!content) {
      showToast('Vui lòng nhập nội dung file .proto', 'warning');
      return;
    }

    try {
      const res = await fetch('/api/grpc/proto/parse', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: content })
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Failed to parse proto content');
      }

      grpcState.services = data.services || [];
      populateGRPCServices(grpcState.services);
      closeGrpcProtoModal();
      showToast(`Đã nạp thành công ${grpcState.services.length} services từ .proto!`, 'success');
    } catch (err) {
      showToast('Parse proto error: ' + err.message, 'danger');
    }
  }

  function setupGRPCResizer() {
    if (!el.grpcSplitResizer) return;
    let isDragging = false;
    let startX = 0;
    let startLeftWidth = 0;

    const reqPane = document.querySelector('.grpc-req-pane');

    el.grpcSplitResizer.addEventListener('mousedown', (e) => {
      isDragging = true;
      startX = e.clientX;
      startLeftWidth = reqPane ? reqPane.getBoundingClientRect().width : 400;
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    });

    window.addEventListener('mousemove', (e) => {
      if (!isDragging || !reqPane) return;
      const dx = e.clientX - startX;
      const newWidth = Math.max(280, Math.min(window.innerWidth - 300, startLeftWidth + dx));
      reqPane.style.flex = `0 0 ${newWidth}px`;
      if (grpcState.editor) grpcState.editor.layout();
      if (grpcState.respEditor) grpcState.respEditor.layout();
    });

    window.addEventListener('mouseup', () => {
      if (isDragging) {
        isDragging = false;
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
        if (grpcState.editor) grpcState.editor.layout();
        if (grpcState.respEditor) grpcState.respEditor.layout();
      }
    });
  }

  function setupGRPCEventListeners() {
    // Mode Switcher buttons
    if (el.btnModeES) {
      el.btnModeES.addEventListener('click', () => switchStudioMode('es'));
    }
    if (el.btnModeGRPC) {
      el.btnModeGRPC.addEventListener('click', () => switchStudioMode('grpc'));
    }

    // Reflect
    if (el.btnGrpcReflect) {
      el.btnGrpcReflect.addEventListener('click', reflectGRPCServices);
    }

    // Service & Method change
    if (el.grpcServiceSelect) {
      el.grpcServiceSelect.addEventListener('change', onGRPCServiceChange);
    }
    if (el.grpcMethodSelect) {
      el.grpcMethodSelect.addEventListener('change', onGRPCMethodChange);
    }

    // Invoke
    if (el.btnGrpcInvoke) {
      el.btnGrpcInvoke.addEventListener('click', invokeGRPC);
    }

    // Body actions
    if (el.btnGrpcFormatBody) {
      el.btnGrpcFormatBody.addEventListener('click', formatGRPCBody);
    }
    if (el.btnGrpcResetMock) {
      el.btnGrpcResetMock.addEventListener('click', resetGRPCMock);
    }

    // Copy Response
    if (el.btnCopyGrpcResponse) {
      el.btnCopyGrpcResponse.addEventListener('click', copyGRPCResponse);
    }

    // Metadata
    if (el.btnAddMetaRow) {
      el.btnAddMetaRow.addEventListener('click', () => addGRPCMetaRow());
    }

    // Proto Modal
    if (el.btnOpenGrpcProtoModal) {
      el.btnOpenGrpcProtoModal.addEventListener('click', openGrpcProtoModal);
    }
    if (el.btnCloseGrpcProtoModal) {
      el.btnCloseGrpcProtoModal.addEventListener('click', closeGrpcProtoModal);
    }
    if (el.btnCancelGrpcProto) {
      el.btnCancelGrpcProto.addEventListener('click', closeGrpcProtoModal);
    }
    if (el.btnParseProtoSubmit) {
      el.btnParseProtoSubmit.addEventListener('click', parseProtoDefinition);
    }

    // Request tabs
    document.querySelectorAll('.grpc-tab').forEach(tab => {
      tab.addEventListener('click', () => {
        document.querySelectorAll('.grpc-tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.grpc-tab-panel').forEach(p => p.classList.remove('active'));
        tab.classList.add('active');
        const targetId = tab.getAttribute('data-tab');
        const panel = document.getElementById(targetId);
        if (panel) panel.classList.add('active');
        if (targetId === 'grpc-tab-body' && grpcState.editor) {
          setTimeout(() => grpcState.editor.layout(), 30);
        }
      });
    });

    // Response tabs
    document.querySelectorAll('.grpc-resp-tab').forEach(tab => {
      tab.addEventListener('click', () => {
        document.querySelectorAll('.grpc-resp-tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.grpc-resp-panel').forEach(p => p.classList.remove('active'));
        tab.classList.add('active');
        const targetId = tab.getAttribute('data-tab');
        const panel = document.getElementById(targetId);
        if (panel) panel.classList.add('active');
        if (targetId === 'grpc-resp-tab-body' && grpcState.respEditor) {
          setTimeout(() => grpcState.respEditor.layout(), 30);
        }
      });
    });

    // Setup Resizer
    setupGRPCResizer();

    // Default metadata row
    addGRPCMetaRow('authorization', '');

    // Global Keydown hook (Cmd+Enter or Ctrl+Enter)
    window.addEventListener('keydown', (e) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        if (grpcState.currentMode === 'grpc') {
          e.preventDefault();
          invokeGRPC();
        }
      }
    });
  }

  document.addEventListener('DOMContentLoaded', init);
})();
