// Common functions for all pages
document.addEventListener('DOMContentLoaded', function() {
    const fileInput = document.getElementById('fileInput');
    const uploadLabel = document.querySelector('.upload-label');
    const submitBtn = document.getElementById('submitBtn');

    if (fileInput && uploadLabel) {
        // File input change handler
        fileInput.addEventListener('change', handleFileSelect);

        // Drag and drop handlers
        uploadLabel.addEventListener('dragover', handleDragOver);
        uploadLabel.addEventListener('dragleave', handleDragLeave);
        uploadLabel.addEventListener('drop', handleDrop);
    }

    // Initialize page-specific handlers
    if (document.getElementById('splitForm')) {
        initSplitPage();
    } else if (document.getElementById('mergeForm')) {
        initMergePage();
    } else if (document.getElementById('compressForm')) {
        initCompressPage();
    } else if (document.getElementById('compressImageForm')) {
        initCompressImagePage();
    } else if (document.getElementById('removePasswordForm')) {
        initRemovePasswordPage();
    } else if (document.getElementById('addPasswordForm')) {
        initAddPasswordPage();
    } else if (document.getElementById('removePageForm')) {
        initRemovePagePage();
    } else if (document.getElementById('imageToPDFForm')) {
        initImageToPDFPage();
    }
});

function handleFileSelect(e) {
    const files = e.target.files;
    if (files.length > 0) {
        updateFileInfo(files);
        document.getElementById('submitBtn').disabled = false;
    }
}

function handleDragOver(e) {
    e.preventDefault();
    e.stopPropagation();
    e.currentTarget.classList.add('drag-over');
}

function handleDragLeave(e) {
    e.preventDefault();
    e.stopPropagation();
    e.currentTarget.classList.remove('drag-over');
}

function handleDrop(e) {
    e.preventDefault();
    e.stopPropagation();
    e.currentTarget.classList.remove('drag-over');

    const files = e.dataTransfer.files;
    if (files.length > 0) {
        document.getElementById('fileInput').files = files;
        updateFileInfo(files);
        document.getElementById('submitBtn').disabled = false;
    }
}

function updateFileInfo(files) {
    const fileInfo = document.getElementById('fileInfo');
    if (fileInfo && files.length === 1) {
        const file = files[0];
        const sizeMB = (file.size / (1024 * 1024)).toFixed(2);
        fileInfo.innerHTML = `<strong>Selected:</strong> ${file.name} (${sizeMB} MB)`;
        fileInfo.style.display = 'block';
    }
}

function showProgress() {
    document.getElementById('progress').style.display = 'block';
    document.getElementById('result').style.display = 'none';
    document.getElementById('submitBtn').disabled = true;
}

function hideProgress() {
    document.getElementById('progress').style.display = 'none';
    document.getElementById('submitBtn').disabled = false;
}

function showResult(message, isError = false, downloadUrl = null) {
    const result = document.getElementById('result');
    result.className = 'result' + (isError ? ' error' : '');

    let html = `<h3>${isError ? 'Error' : 'Success'}!</h3><p>${message}</p>`;

    if (downloadUrl) {
        html += `<a href="${downloadUrl}" class="download-link" download>Download Result</a>`;
    }

    result.innerHTML = html;
    result.style.display = 'block';
    hideProgress();
}

// Split page
function initSplitPage() {
    const form = document.getElementById('splitForm');
    const rangeRadio = document.getElementById('customRange');
    const pageRangeInput = document.getElementById('pageRange');

    // Enable/disable page range input
    document.querySelectorAll('input[name="splitMode"]').forEach(radio => {
        radio.addEventListener('change', function() {
            pageRangeInput.disabled = !rangeRadio.checked;
        });
    });

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const formData = new FormData(form);
        const splitMode = formData.get('splitMode');

        if (splitMode === 'range' && !formData.get('pageRange')) {
            showResult('Please enter page ranges', true);
            return;
        }

        showProgress();

        try {
            const response = await fetch('/api/split', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                showResult(data.message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Split failed', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}

// Merge page
function initMergePage() {
    const form = document.getElementById('mergeForm');
    const fileInput = document.getElementById('fileInput');
    const fileList = document.getElementById('fileList');
    const fileItems = document.getElementById('fileItems');
    let selectedFiles = [];

    fileInput.addEventListener('change', function(e) {
        selectedFiles = Array.from(e.target.files);
        updateFileList();
    });

    function updateFileList() {
        if (selectedFiles.length < 2) {
            fileList.style.display = 'none';
            document.getElementById('submitBtn').disabled = true;
            return;
        }

        fileList.style.display = 'block';
        document.getElementById('submitBtn').disabled = false;

        fileItems.innerHTML = selectedFiles.map((file, index) => `
            <div class="file-item" draggable="true" data-index="${index}">
                <span class="file-item-name">${index + 1}. ${file.name}</span>
                <span class="file-item-remove" onclick="removeFile(${index})">✕</span>
            </div>
        `).join('');

        // Add drag and drop for reordering
        const items = fileItems.querySelectorAll('.file-item');
        items.forEach(item => {
            item.addEventListener('dragstart', handleDragStart);
            item.addEventListener('dragover', handleDragOverItem);
            item.addEventListener('drop', handleDropItem);
            item.addEventListener('dragend', handleDragEnd);
        });
    }

    window.removeFile = function(index) {
        selectedFiles.splice(index, 1);
        updateFileList();
    };

    let draggedIndex;

    function handleDragStart(e) {
        draggedIndex = parseInt(e.currentTarget.dataset.index);
        e.currentTarget.classList.add('dragging');
    }

    function handleDragOverItem(e) {
        e.preventDefault();
    }

    function handleDropItem(e) {
        e.preventDefault();
        const dropIndex = parseInt(e.currentTarget.dataset.index);

        if (draggedIndex !== dropIndex) {
            const draggedFile = selectedFiles[draggedIndex];
            selectedFiles.splice(draggedIndex, 1);
            selectedFiles.splice(dropIndex, 0, draggedFile);
            updateFileList();
        }
    }

    function handleDragEnd(e) {
        e.currentTarget.classList.remove('dragging');
    }

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        if (selectedFiles.length < 2) {
            showResult('Please select at least 2 PDF files', true);
            return;
        }

        showProgress();

        const formData = new FormData();
        selectedFiles.forEach(file => {
            formData.append('files', file);
        });

        try {
            const response = await fetch('/api/merge', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                showResult(data.message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Merge failed', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}

// Compress page
function initCompressPage() {
    const form = document.getElementById('compressForm');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        showProgress();

        const formData = new FormData(form);

        try {
            const response = await fetch('/api/compress', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                let message = data.message;
                if (data.originalSize && data.compressedSize) {
                    const reduction = ((1 - data.compressedSize / data.originalSize) * 100).toFixed(1);
                    message += ` Size reduced by ${reduction}%`;
                }
                showResult(message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Compression failed', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}
// Compress image page
function initCompressImagePage() {
    const form = document.getElementById('compressImageForm');
    const fileInput = document.getElementById('fileInput');
    const qualitySlider = document.getElementById('quality');
    const qualityValue = document.getElementById('qualityValue');
    const resizeCheckbox = document.getElementById('resize');
    const resizeOptions = document.getElementById('resizeOptions');
    const previewDiv = document.getElementById('imagePreview');
    const previewImg = document.getElementById('previewImg');

    // Update quality value display
    if (qualitySlider && qualityValue) {
        qualitySlider.addEventListener('input', function() {
            qualityValue.textContent = this.value;
        });
    }

    // Toggle resize options
    if (resizeCheckbox && resizeOptions) {
        resizeCheckbox.addEventListener('change', function() {
            resizeOptions.style.display = this.checked ? 'block' : 'none';
        });
    }

    // Handle dimension presets
    const presetSelect = document.getElementById('preset');
    const targetWidth = document.getElementById('targetWidth');
    const targetHeight = document.getElementById('targetHeight');
    const resizeMode = document.getElementById('resizeMode');

    if (presetSelect && targetWidth && targetHeight) {
        presetSelect.addEventListener('change', function() {
            const presets = {
                'passport': { width: 600, height: 600, mode: 'exact' },
                'id': { width: 450, height: 600, mode: 'exact' },
                'hd': { width: 1920, height: 1080, mode: 'max' },
                'square-1024': { width: 1024, height: 1024, mode: 'exact' },
                'square-512': { width: 512, height: 512, mode: 'exact' }
            };

            const preset = presets[this.value];
            if (preset) {
                targetWidth.value = preset.width;
                targetHeight.value = preset.height;
                resizeMode.value = preset.mode;
            } else {
                // Custom - clear values
                targetWidth.value = '';
                targetHeight.value = '';
            }
        });
    }

    // Show image preview
    if (fileInput && previewDiv && previewImg) {
        fileInput.addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file && file.type.startsWith('image/')) {
                const reader = new FileReader();
                reader.onload = function(e) {
                    previewImg.src = e.target.result;
                    previewDiv.style.display = 'block';
                };
                reader.readAsDataURL(file);
            }
        });
    }

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        showProgress();

        const formData = new FormData(form);

        try {
            const response = await fetch('/api/compress-image', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                let message = data.message;
                if (data.originalSize && data.compressedSize) {
                    const reduction = ((1 - data.compressedSize / data.originalSize) * 100).toFixed(1);
                    const originalMB = (data.originalSize / (1024 * 1024)).toFixed(2);
                    const compressedMB = (data.compressedSize / (1024 * 1024)).toFixed(2);
                    message += `<br>Original: ${originalMB} MB → Compressed: ${compressedMB} MB (${reduction}% reduction)`;
                }
                showResult(message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Compression failed', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}

// Remove password page
function initRemovePasswordPage() {
    const form = document.getElementById('removePasswordForm');
    const passwordInput = document.getElementById('passwordInput');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const password = passwordInput.value.trim();
        if (!password) {
            showResult('Please enter the PDF password', true);
            return;
        }

        showProgress();

        const formData = new FormData(form);

        try {
            const response = await fetch('/api/remove-password', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                showResult(data.message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Failed to remove password', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}

// Add password page
function initAddPasswordPage() {
    const form = document.getElementById('addPasswordForm');
    const passwordInput = document.getElementById('passwordInput');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const password = passwordInput.value.trim();
        if (!password) {
            showResult('Please enter a password', true);
            return;
        }

        showProgress();

        const formData = new FormData(form);

        try {
            const response = await fetch('/api/add-password', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                showResult(data.message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Failed to add password', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}

// Remove page page
function initRemovePagePage() {
    const form = document.getElementById('removePageForm');
    const pageRangeInput = document.getElementById('pageRangeInput');

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        const pageRange = pageRangeInput.value.trim();
        if (!pageRange) {
            showResult('Please enter page numbers to remove', true);
            return;
        }

        showProgress();

        const formData = new FormData(form);

        try {
            const response = await fetch('/api/remove-page', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                showResult(data.message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Failed to remove pages', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}

// Images to PDF page
function initImageToPDFPage() {
    const form = document.getElementById('imageToPDFForm');
    const fileInput = document.getElementById('fileInput');
    const fileList = document.getElementById('fileList');
    const fileItems = document.getElementById('fileItems');
    let selectedFiles = [];

    fileInput.addEventListener('change', function(e) {
        selectedFiles = Array.from(e.target.files);
        updateFileList();
    });

    function updateFileList() {
        if (selectedFiles.length === 0) {
            fileList.style.display = 'none';
            document.getElementById('submitBtn').disabled = true;
            return;
        }

        fileList.style.display = 'block';
        document.getElementById('submitBtn').disabled = false;

        fileItems.innerHTML = selectedFiles.map((file, index) => `
            <div class="file-item" draggable="true" data-index="${index}">
                <span class="file-item-name">${index + 1}. ${file.name}</span>
                <span class="file-item-remove" onclick="removeImageFile(${index})">✕</span>
            </div>
        `).join('');

        // Add drag and drop for reordering
        const items = fileItems.querySelectorAll('.file-item');
        items.forEach(item => {
            item.addEventListener('dragstart', handleDragStart);
            item.addEventListener('dragover', handleDragOverItem);
            item.addEventListener('drop', handleDropItem);
            item.addEventListener('dragend', handleDragEnd);
        });
    }

    window.removeImageFile = function(index) {
        selectedFiles.splice(index, 1);
        updateFileList();
    };

    let draggedIndex;

    function handleDragStart(e) {
        draggedIndex = parseInt(e.currentTarget.dataset.index);
        e.currentTarget.classList.add('dragging');
    }

    function handleDragOverItem(e) {
        e.preventDefault();
    }

    function handleDropItem(e) {
        e.preventDefault();
        const dropIndex = parseInt(e.currentTarget.dataset.index);

        if (draggedIndex !== dropIndex) {
            const draggedFile = selectedFiles[draggedIndex];
            selectedFiles.splice(draggedIndex, 1);
            selectedFiles.splice(dropIndex, 0, draggedFile);
            updateFileList();
        }
    }

    function handleDragEnd(e) {
        e.currentTarget.classList.remove('dragging');
    }

    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        if (selectedFiles.length === 0) {
            showResult('Please select at least 1 image file', true);
            return;
        }

        showProgress();

        const formData = new FormData();
        selectedFiles.forEach(file => {
            formData.append('files', file);
        });

        try {
            const response = await fetch('/api/image-to-pdf', {
                method: 'POST',
                body: formData
            });

            const data = await response.json();

            if (response.ok) {
                showResult(data.message, false, data.downloadUrl);
            } else {
                showResult(data.error || 'Conversion failed', true);
            }
        } catch (error) {
            showResult('Network error: ' + error.message, true);
        }
    });
}
