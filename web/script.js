const API_BASE = 'http://localhost:8080/api/v1';

const shortenForm = document.getElementById('shortenForm');
const lookupForm = document.getElementById('lookupForm');
const resultDiv = document.getElementById('result');
const lookupResultDiv = document.getElementById('lookupResult');
const redirectResultDiv = document.getElementById('redirectResult');
const apiStatus = document.getElementById('apiStatus');
const dbStatus = document.getElementById('dbStatus');

document.addEventListener('DOMContentLoaded', function() {
    checkHealth();
});

function showResult(element, message, type = 'success') {
    element.innerHTML = message;
    element.className = `result ${type}`;
}

function showError(element, message) {
    showResult(element, message, 'error');
}

function showInfo(element, message) {
    showResult(element, message, 'info');
}

shortenForm.addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const urlInput = document.getElementById('urlInput');
    const url = urlInput.value.trim();
    
    if (!url) {
        showError(resultDiv, 'Пожалуйста, введите URL');
        return;
    }
    
    try {
        showInfo(resultDiv, 'Создаем короткую ссылку...');
        
        const response = await fetch(`${API_BASE}/shorten`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: url })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        
        const shortUrl = `http://localhost:8080/${data.short_url}`;
        const resultHTML = `
            <div class="success-url">
                <div>
                    <strong>✅ Короткая ссылка создана!</strong><br>
                    <a href="${shortUrl}" target="_blank" class="short-url-link">${shortUrl}</a>
                </div>
                <button class="copy-btn" onclick="copyToClipboard('${shortUrl}')">Копировать</button>
            </div>
            <div class="url-display">
                <strong>Оригинальный URL:</strong><br>
                <a href="${url}" target="_blank">${url}</a>
            </div>
            <div class="test-redirect">
                <button onclick="testRedirect('${data.short_url}')" class="test-btn">🔗 Протестировать редирект</button>
            </div>
        `;
        
        showResult(resultDiv, resultHTML);
        urlInput.value = '';
        
    } catch (error) {
        console.error('Error:', error);
        showError(resultDiv, `Ошибка: ${error.message}`);
    }
});

lookupForm.addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const shortCodeInput = document.getElementById('shortCodeInput');
    const shortCode = shortCodeInput.value.trim();
    
    if (!shortCode) {
        showError(lookupResultDiv, 'Пожалуйста, введите короткий код');
        return;
    }
    
    try {
        showInfo(lookupResultDiv, 'Ищем информацию о ссылке...');
        
        const response = await fetch(`${API_BASE}/url/${shortCode}`);
        
        if (!response.ok) {
            if (response.status === 404) {
                throw new Error('Ссылка не найдена');
            }
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        
        const shortUrl = `http://localhost:8080/${data.short_url}`;
        const resultHTML = `
            <div class="url-display">
                <strong>Короткая ссылка:</strong><br>
                <a href="${shortUrl}" target="_blank" class="short-url-link">${shortUrl}</a>
            </div>
            <div class="url-display">
                <strong>Оригинальный URL:</strong><br>
                <a href="${data.original_url}" target="_blank">${data.original_url}</a>
            </div>
            <div class="url-info">
                <strong>Статистика:</strong><br>
                • Кликов: ${data.clicks || 0}<br>
                • Создана: ${new Date(data.created_at).toLocaleString('ru-RU')}
            </div>
            <div class="action-buttons">
                <button class="copy-btn" onclick="copyToClipboard('${shortUrl}')">Копировать короткую ссылку</button>
                <button class="test-btn" onclick="testRedirect('${data.short_url}')">🔗 Тестировать редирект</button>
            </div>
        `;
        
        showResult(lookupResultDiv, resultHTML);
        shortCodeInput.value = '';
        
    } catch (error) {
        console.error('Error:', error);
        showError(lookupResultDiv, `Ошибка: ${error.message}`);
    }
});

// Функция для тестирования редиректа
async function testRedirect(shortCode) {
    try {
        showInfo(redirectResultDiv, 'Тестируем редирект...');
        
        const response = await fetch(`http://localhost:8080/${shortCode}`, {
            method: 'GET',
            redirect: 'manual'
        });
        
        if (response.status >= 300 && response.status < 400) {
            const redirectUrl = response.headers.get('Location');
            showResult(redirectResultDiv, `
                <strong>✅ Редирект работает!</strong><br>
                <strong>Короткая ссылка:</strong> http://localhost:8080/${shortCode}<br>
                <strong>Перенаправление на:</strong> <a href="${redirectUrl}" target="_blank">${redirectUrl}</a><br>
                <strong>Статус:</strong> ${response.status}
            `);
        } else {
            showError(redirectResultDiv, `Неожиданный статус: ${response.status}`);
        }
        
    } catch (error) {
        console.error('Error:', error);
        showError(redirectResultDiv, `Ошибка: ${error.message}`);
    }
}

// Тестирование редиректа из поля ввода
async function testRedirectFromInput() {
    const redirectInput = document.getElementById('redirectInput');
    const shortCode = redirectInput.value.trim();
    
    if (!shortCode) {
        showError(redirectResultDiv, 'Пожалуйста, введите короткий код');
        return;
    }
    
    await testRedirect(shortCode);
}

async function checkHealth() {
    apiStatus.textContent = 'Проверка...';
    apiStatus.className = 'status checking';
    dbStatus.textContent = 'Проверка...';
    dbStatus.className = 'status checking';
    
    try {
        const response = await fetch('http://localhost:8080/health');
        if (response.ok) {
            apiStatus.textContent = 'Онлайн';
            apiStatus.className = 'status online';
            dbStatus.textContent = 'Онлайн';
            dbStatus.className = 'status online';
        } else {
            throw new Error(`Status: ${response.status}`);
        }
    } catch (error) {
        console.error('Health check failed:', error);
        apiStatus.textContent = 'Оффлайн';
        apiStatus.className = 'status offline';
        dbStatus.textContent = 'Оффлайн';
        dbStatus.className = 'status offline';
    }
}

async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        alert('✅ Скопировано в буфер обмена!');
    } catch (err) {
        const textArea = document.createElement('textarea');
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        alert('✅ Скопировано в буфер обмена!');
    }
}

setInterval(checkHealth, 30000);