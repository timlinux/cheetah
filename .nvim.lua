-- SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
-- SPDX-License-Identifier: MIT

-- Cheetah project-specific Neovim configuration
-- This file is automatically loaded when opening Neovim in the project directory

-- Set Go-specific options
vim.opt_local.tabstop = 4
vim.opt_local.shiftwidth = 4
vim.opt_local.expandtab = false

-- Project root detection
vim.g.cheetah_root = vim.fn.getcwd()

-- Test output parsing for quickfix
-- Go test output format: filename:line: error message
vim.opt_local.errorformat = vim.opt_local.errorformat + '%f:%l:%c: %m'
vim.opt_local.errorformat = vim.opt_local.errorformat + '%f:%l: %m'
vim.opt_local.errorformat = vim.opt_local.errorformat + '--- FAIL: %m'
vim.opt_local.errorformat = vim.opt_local.errorformat + 'FAIL%m'

-- Function to run Go tests and populate quickfix
local function run_go_tests(package)
  package = package or './...'
  vim.cmd('copen')
  vim.cmd('setlocal modifiable')
  vim.fn.setqflist({}, 'r')

  local cmd = 'go test -v ' .. package .. ' 2>&1'
  vim.cmd('cexpr system("' .. cmd .. '")')
  vim.cmd('copen')
end

-- Function to run current package tests
local function run_current_package_tests()
  local file = vim.fn.expand('%:p')
  local dir = vim.fn.fnamemodify(file, ':h')
  run_go_tests('./' .. vim.fn.fnamemodify(dir, ':t'))
end

-- Function to run test under cursor
local function run_test_under_cursor()
  local line = vim.fn.getline('.')
  local test_name = line:match('func (Test%w+)')
  if test_name then
    local file = vim.fn.expand('%:p')
    local dir = vim.fn.fnamemodify(file, ':h')
    local pkg = './' .. vim.fn.fnamemodify(dir, ':t')
    vim.cmd('copen')
    local cmd = 'go test -v -run ' .. test_name .. ' ' .. pkg .. ' 2>&1'
    vim.cmd('cexpr system("' .. cmd .. '")')
  else
    vim.notify('No test function found under cursor', vim.log.levels.WARN)
  end
end

-- Function to run Playwright tests
local function run_playwright_tests()
  vim.cmd('copen')
  vim.cmd('terminal cd web && npm run test')
end

-- Function to run all tests (Go + Playwright)
local function run_all_tests()
  vim.cmd('copen')
  vim.fn.setqflist({}, 'r')
  vim.cmd('cexpr system("make test 2>&1")')
end

-- Custom commands for this project
vim.api.nvim_create_user_command('CheetahBuild', '!make build', {})
vim.api.nvim_create_user_command('CheetahRun', '!make run', {})
vim.api.nvim_create_user_command('CheetahTest', function() run_all_tests() end, {})
vim.api.nvim_create_user_command('CheetahTestGo', function() run_go_tests() end, {})
vim.api.nvim_create_user_command('CheetahTestPkg', function() run_current_package_tests() end, {})
vim.api.nvim_create_user_command('CheetahTestCursor', function() run_test_under_cursor() end, {})
vim.api.nvim_create_user_command('CheetahTestWeb', function() run_playwright_tests() end, {})
vim.api.nvim_create_user_command('CheetahFmt', '!make fmt', {})
vim.api.nvim_create_user_command('CheetahLint', '!make lint', {})
vim.api.nvim_create_user_command('CheetahDocsDev', '!make docs-dev &', {})
vim.api.nvim_create_user_command('CheetahDocsBuild', '!make docs-build', {})
vim.api.nvim_create_user_command('CheetahDocsOpen', '!make docs-open', {})
vim.api.nvim_create_user_command('CheetahWebDev', '!make web-dev', {})

-- Which-key integration for project keybindings
local ok, wk = pcall(require, "which-key")
if ok then
  wk.register({
    p = {
      name = "Project (Cheetah)",
      -- Build & Run
      b = { "<cmd>!make build<cr>", "Build binary" },
      r = { "<cmd>!make run<cr>", "Run application" },

      -- Testing (with quickfix support)
      t = {
        name = "Tests",
        t = { "<cmd>CheetahTest<cr>", "Run all tests (quickfix)" },
        g = { "<cmd>CheetahTestGo<cr>", "Run Go tests (quickfix)" },
        p = { "<cmd>CheetahTestPkg<cr>", "Run current package tests" },
        c = { "<cmd>CheetahTestCursor<cr>", "Run test under cursor" },
        w = { "<cmd>CheetahTestWeb<cr>", "Run Playwright tests" },
        n = { "<cmd>cnext<cr>", "Next test failure" },
        N = { "<cmd>cprev<cr>", "Previous test failure" },
        o = { "<cmd>copen<cr>", "Open quickfix list" },
        q = { "<cmd>cclose<cr>", "Close quickfix list" },
      },

      -- Server/Client
      s = { "<cmd>!make server<cr>", "Start server only" },
      c = { "<cmd>!make client<cr>", "Start client only" },
      S = { "<cmd>!make start-backend<cr>", "Start backend (background)" },
      K = { "<cmd>!make stop-backend<cr>", "Stop backend" },

      -- Code Quality
      f = { "<cmd>!make fmt<cr>", "Format code" },
      l = { "<cmd>!make lint<cr>", "Lint code" },

      -- Documentation (Hugo)
      d = {
        name = "Documentation",
        d = { "<cmd>!make docs-dev &<cr>", "Start Hugo dev server" },
        b = { "<cmd>!make docs-build<cr>", "Build documentation" },
        c = { "<cmd>!make docs-clean<cr>", "Clean documentation" },
        o = { "<cmd>!make docs-open<cr>", "Open docs in browser" },
        n = { "<cmd>terminal make docs-new<cr>", "New documentation page" },
      },

      -- Web Frontend
      w = {
        name = "Web Frontend",
        i = { "<cmd>!make web-install<cr>", "Install web dependencies" },
        d = { "<cmd>!make web-dev<cr>", "Start web dev server" },
        b = { "<cmd>!make web-build<cr>", "Build web for production" },
        s = { "<cmd>!make web-start<cr>", "Start backend + web" },
        t = { "<cmd>CheetahTestWeb<cr>", "Run Playwright tests" },
      },

      -- Nix
      n = {
        name = "Nix",
        b = { "<cmd>!nix build<cr>", "Nix build" },
        r = { "<cmd>!nix run<cr>", "Nix run" },
        d = { "<cmd>!nix run .#docs-serve<cr>", "Nix docs serve" },
      },

      -- Quickfix navigation shortcuts
      q = {
        name = "Quickfix",
        o = { "<cmd>copen<cr>", "Open quickfix" },
        c = { "<cmd>cclose<cr>", "Close quickfix" },
        n = { "<cmd>cnext<cr>", "Next item" },
        p = { "<cmd>cprev<cr>", "Previous item" },
        f = { "<cmd>cfirst<cr>", "First item" },
        l = { "<cmd>clast<cr>", "Last item" },
      },
    },
  }, { prefix = "<leader>" })
end

-- Autocommands for this project
local cheetah_group = vim.api.nvim_create_augroup("CheetahProject", { clear = true })

-- Format Go files on save
vim.api.nvim_create_autocmd("BufWritePre", {
  group = cheetah_group,
  pattern = "*.go",
  callback = function()
    vim.lsp.buf.format({ async = false })
  end,
})

-- Set filetype for Hugo templates
vim.api.nvim_create_autocmd({ "BufRead", "BufNewFile" }, {
  group = cheetah_group,
  pattern = vim.g.cheetah_root .. "/hugo/themes/cheetah/layouts/**/*.html",
  callback = function()
    vim.bo.filetype = "html.gotmpl"
  end,
})

-- Auto-open quickfix on test failure
vim.api.nvim_create_autocmd("QuickFixCmdPost", {
  group = cheetah_group,
  pattern = "[^l]*",
  callback = function()
    vim.cmd('copen')
  end,
})

-- Print confirmation
vim.notify("Cheetah project configuration loaded (with test support)", vim.log.levels.INFO)
