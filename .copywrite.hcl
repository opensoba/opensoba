schema_version = 1

project {
  license        = "MPL-2.0"
  copyright_year = 2013

  # (OPTIONAL) A list of globs that should not have copyright/license headers.
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  header_ignore = [
    # Forked and modified UI libs
    "ui/packages/sabo-ui/app/utils/dom/event-target/**",
    "ui/packages/sabo-ui/lib/rehype-prism/**",
    "ui/packages/sabo-ui/lib/block-slots/**",

    # UI file that do not render properly with copyright headers
    "ui/packages/sabo-ui/app/components/brand-loader/enterprise.hbs",
    "ui/packages/sabo-ui/app/components/brand-loader/index.hbs",

    # ignore specific test data files
    "agent/uiserver/testdata/**",

    # generated files 
    "agent/structs/structs.deepcopy.go",
    "agent/proxycfg/proxycfg.deepcopy.go",
    "agent/grpc-middleware/rate_limit_mappings.gen.go",
    "agent/uiserver/dist/**",
  ]
}
