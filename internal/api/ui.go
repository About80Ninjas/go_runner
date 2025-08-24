// internal/api/ui.go
package api

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templates embed.FS

//go:embed static/*
var static embed.FS

// adminUIHandler serves the admin UI
func (s *Server) adminUIHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Binary Executor Admin</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; background: #f5f5f5; }
        .header { background: #2c3e50; color: white; padding: 1rem 2rem; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 1rem; }
        .card { background: white; border-radius: 8px; padding: 1.5rem; margin-bottom: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .card h2 { margin-bottom: 1rem; color: #2c3e50; }
        .form-group { margin-bottom: 1rem; }
        .form-group label { display: block; margin-bottom: 0.5rem; font-weight: 500; }
        .form-group input, .form-group select { width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
        .btn { padding: 0.5rem 1rem; border: none; border-radius: 4px; cursor: pointer; font-weight: 500; transition: opacity 0.2s; }
        .btn:hover { opacity: 0.9; }
        .btn-primary { background: #3498db; color: white; }
        .btn-success { background: #27ae60; color: white; }
        .btn-danger { background: #e74c3c; color: white; }
        .binary-list { display: grid; gap: 1rem; }
        .binary-item { display: flex; justify-content: space-between; align-items: center; padding: 1rem; background: #f8f9fa; border-radius: 4px; }
        .status { padding: 0.25rem 0.5rem; border-radius: 3px; font-size: 0.875rem; font-weight: 500; }
        .status-ready { background: #d4edda; color: #155724; }
        .status-building { background: #fff3cd; color: #856404; }
        .status-failed { background: #f8d7da; color: #721c24; }
        .status-pending { background: #d1ecf1; color: #0c5460; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Binary Executor Admin Panel</h1>
    </div>
    <div class="container">
        <div class="card">
            <h2>Add New Binary</h2>
            <form id="addBinaryForm">
                <div class="form-group">
                    <label>Name</label>
                    <input type="text" name="name" required>
                </div>
                <div class="form-group">
                    <label>Description</label>
                    <input type="text" name="description">
                </div>
                <div class="form-group">
                    <label>Repository URL</label>
                    <input type="url" name="repo_url" placeholder="https://github.com/user/repo.git" required>
                </div>
                <div class="form-group">
                    <label>Branch</label>
                    <input type="text" name="branch" value="main" required>
                </div>
                <div class="form-group">
                    <label>Build Path (relative to repo root)</label>
                    <input type="text" name="build_path" placeholder="./cmd/app" value=".">
                </div>
                <button type="submit" class="btn btn-primary">Add Binary</button>
            </form>
        </div>

        <div class="card">
            <h2>Managed Binaries</h2>
            <div id="binaryList" class="binary-list"></div>
        </div>
    </div>

    <script>
        const API_BASE = '/api/v1';
        const token = new URLSearchParams(window.location.search).get('token') || localStorage.getItem('admin_token');
        
        if (!token) {
            const inputToken = prompt('Enter admin token:');
            if (inputToken) {
                localStorage.setItem('admin_token', inputToken);
                window.location.href = window.location.pathname + '?token=' + inputToken;
            }
        }

        async function fetchBinaries() {
            try {
                const response = await fetch(API_BASE + '/binaries', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                const binaries = await response.json();
                displayBinaries(binaries);
            } catch (error) {
                console.error('Failed to fetch binaries:', error);
            }
        }

        function displayBinaries(binaries) {
            const list = document.getElementById('binaryList');
            if (!binaries || binaries.length === 0) {
                list.innerHTML = '<p>No binaries configured</p>';
                return;
            }
            
            list.innerHTML = binaries.map(binary => ` + "`" + `
                <div class="binary-item">
                    <div>
                        <strong>${binary.name}</strong>
                        <small>${binary.description || ''}</small>
                        <div>
                            <small>${binary.repo_url}</small>
                            <span class="status status-${binary.status}">${binary.status}</span>
                        </div>
                    </div>
                    <div>
                        <button class="btn btn-success" onclick="buildBinary('${binary.id}')">Build</button>
                        <button class="btn btn-danger" onclick="deleteBinary('${binary.id}')">Delete</button>
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        async function buildBinary(id) {
            try {
                await fetch(API_BASE + '/binaries/' + id + '/build', {
                    method: 'POST',
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                alert('Build started');
                fetchBinaries();
            } catch (error) {
                alert('Failed to start build: ' + error);
            }
        }

        async function deleteBinary(id) {
            if (!confirm('Delete this binary?')) return;
            try {
                await fetch(API_BASE + '/binaries/' + id, {
                    method: 'DELETE',
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                fetchBinaries();
            } catch (error) {
                alert('Failed to delete: ' + error);
            }
        }

        document.getElementById('addBinaryForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData);
            
            try {
                await fetch(API_BASE + '/binaries', {
                    method: 'POST',
                    headers: {
                        'Authorization': 'Bearer ' + token,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(data)
                });
                e.target.reset();
                fetchBinaries();
            } catch (error) {
                alert('Failed to add binary: ' + error);
            }
        });

        fetchBinaries();
        setInterval(fetchBinaries, 5000);
    </script>
</body>
</html>`

	t, _ := template.New("admin").Parse(tmpl)
	t.Execute(w, nil)
}
