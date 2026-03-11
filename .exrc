" Cheetah project Neovim keybindings
" All project-specific shortcuts are under <leader>p

" Ensure which-key is available for descriptions
if exists(':WhichKey')
  " Register the p group
  lua << EOF
  local wk = require("which-key")
  wk.register({
    p = {
      name = "Project (Cheetah)",
      -- Build & Run
      b = { "<cmd>!make build<cr>", "Build binary" },
      r = { "<cmd>!make run<cr>", "Run application" },
      t = { "<cmd>!make test<cr>", "Run tests" },

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
      },

      -- Nix
      n = {
        name = "Nix",
        b = { "<cmd>!nix build<cr>", "Nix build" },
        r = { "<cmd>!nix run<cr>", "Nix run" },
        d = { "<cmd>!nix run .#docs-serve<cr>", "Nix docs serve" },
      },
    },
  }, { prefix = "<leader>" })
EOF
else
  " Fallback without which-key
  " Build & Run
  nnoremap <leader>pb :!make build<CR>
  nnoremap <leader>pr :!make run<CR>
  nnoremap <leader>pt :!make test<CR>

  " Server/Client
  nnoremap <leader>ps :!make server<CR>
  nnoremap <leader>pc :!make client<CR>
  nnoremap <leader>pS :!make start-backend<CR>
  nnoremap <leader>pK :!make stop-backend<CR>

  " Code Quality
  nnoremap <leader>pf :!make fmt<CR>
  nnoremap <leader>pl :!make lint<CR>

  " Documentation
  nnoremap <leader>pdd :!make docs-dev &<CR>
  nnoremap <leader>pdb :!make docs-build<CR>
  nnoremap <leader>pdc :!make docs-clean<CR>
  nnoremap <leader>pdo :!make docs-open<CR>
  nnoremap <leader>pdn :terminal make docs-new<CR>

  " Web Frontend
  nnoremap <leader>pwi :!make web-install<CR>
  nnoremap <leader>pwd :!make web-dev<CR>
  nnoremap <leader>pwb :!make web-build<CR>
  nnoremap <leader>pws :!make web-start<CR>

  " Nix
  nnoremap <leader>pnb :!nix build<CR>
  nnoremap <leader>pnr :!nix run<CR>
  nnoremap <leader>pnd :!nix run .#docs-serve<CR>
endif
