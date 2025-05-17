// Store the list ID for API calls
const listId = document.getElementById('listIdField').value;

// Debug on page load
console.log("List ID at page load:", listId);

// Drag and drop functionality
function drag(event) {
    event.dataTransfer.setData("taskId", event.target.getAttribute("data-task-id"));
    event.target.classList.add("dragging");
}

function allowDrop(event) {
    event.preventDefault();
}

function drop(event) {
    event.preventDefault();
    
    // Get the task ID from the dragged element
    const taskId = event.dataTransfer.getData("taskId");
    const draggedElement = document.querySelector(`[data-task-id="${taskId}"]`);
    
    // Remove the dragging class
    draggedElement.classList.remove("dragging");
    
    // Determine the target column
    let targetColumn = event.target;
    while (targetColumn && !targetColumn.classList.contains("column-content")) {
        targetColumn = targetColumn.parentElement;
    }
    
    if (!targetColumn) return;
    
    // Get the column name
    const columnName = targetColumn.parentElement.getAttribute("data-column");
    
    // Show loading overlay
    document.getElementById("loadingOverlay").style.display = "flex";
    
    // Debug the values
    console.log("About to send update request:");
    console.log("- List ID:", listId);
    console.log("- Task ID:", taskId);
    console.log("- Column Name:", columnName);
    
    // Send the update to the server
    updateTaskStatus(taskId, columnName)
        .then((success) => {
            if (success) {
                // Check for and remove the "no-tasks" message if it exists
                const noTasksMessage = targetColumn.querySelector(".no-tasks");
                if (noTasksMessage) {
                    noTasksMessage.remove();
                }
                
                // If successful, move the task to the new column
                targetColumn.appendChild(draggedElement);
                
                // Update the UI based on the new column
                if (columnName === "Done") {
                    draggedElement.classList.add("completed");
                    
                    // Remove "Doing" category tag
                    const doingTags = draggedElement.querySelectorAll(".category-tag");
                    doingTags.forEach(tag => {
                        if (tag.textContent.toLowerCase() === "doing") {
                            tag.remove();
                        }
                    });
                } else if (columnName === "Doing") {
                    draggedElement.classList.remove("completed");
                    
                    // Add "Doing" category tag if it doesn't exist
                    let hasDoingTag = false;
                    const categories = draggedElement.querySelector(".task-categories");
                    const tags = categories ? categories.querySelectorAll(".category-tag") : [];
                    
                    tags.forEach(tag => {
                        if (tag.textContent.toLowerCase() === "doing") {
                            hasDoingTag = true;
                        }
                    });
                    
                    if (!hasDoingTag) {
                        if (!categories) {
                            // Create categories div if it doesn't exist
                            const categoriesDiv = document.createElement("div");
                            categoriesDiv.className = "task-categories";
                            draggedElement.appendChild(categoriesDiv);
                            
                            const doingTag = document.createElement("span");
                            doingTag.className = "category-tag";
                            doingTag.textContent = "Doing";
                            categoriesDiv.appendChild(doingTag);
                        } else {
                            const doingTag = document.createElement("span");
                            doingTag.className = "category-tag";
                            doingTag.textContent = "Doing";
                            categories.appendChild(doingTag);
                        }
                    }
                } else if (columnName === "Not Started") {
                    draggedElement.classList.remove("completed");
                    
                    // Remove "Doing" category tag
                    const doingTags = draggedElement.querySelectorAll(".category-tag");
                    doingTags.forEach(tag => {
                        if (tag.textContent.toLowerCase() === "doing") {
                            tag.remove();
                        }
                    });
                }
                
                // Show success message
                showToast("Task moved successfully!", "success");
            } else {
                showToast("Failed to update task. Please try again.", "error");
            }
        })
        .catch(error => {
            console.error("Error updating task:", error);
            showToast("Error: " + error.message, "error");
        })
        .finally(() => {
            // Hide loading overlay
            document.getElementById("loadingOverlay").style.display = "none";
        });
}

// Function to update task status on the server
async function updateTaskStatus(taskId, columnName) {
    try {
        // Use URL encoded form data instead of FormData
        const params = new URLSearchParams();
        params.append("listId", listId);
        params.append("taskId", taskId);
        params.append("column", columnName);
        
        // Debug logs
        console.log("Sending request with:");
        console.log("listId:", listId);
        console.log("taskId:", taskId);
        console.log("column:", columnName);
        
        const response = await fetch("/api/updateTask", {
            method: "POST",
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: params.toString(),
            credentials: "same-origin"
        });
        
        if (!response.ok) {
            // Get the error text from the response
            const errorText = await response.text();
            console.error("Server error:", errorText);
            throw new Error(errorText || "Server error");
        }
        
        return response.ok;
    } catch (error) {
        console.error("Error updating task:", error);
        return false;
    }
}

// Show toast notification
function showToast(message, type) {
    const toast = document.getElementById("toast");
    toast.textContent = message;
    toast.className = "toast " + type;
    toast.style.display = "block";
    
    // Hide after 3 seconds
    setTimeout(() => {
        toast.style.display = "none";
    }, 3000);
}

// Toggle task importance
function toggleImportance(event, taskCard) {
    // Prevent event from propagating to parent elements (avoid triggering drag)
    event.stopPropagation();
    
    if (!taskCard) return;
    
    // Get task ID and current importance state from the task card
    const taskId = taskCard.getAttribute('data-task-id');
    const currentState = taskCard.getAttribute('data-importance') === 'true';
    
    // Show loading overlay
    document.getElementById("loadingOverlay").style.display = "flex";
    
    // New importance state is the opposite of the current state
    const newImportance = !currentState;
    
    // Debug logs
    console.log("Toggling importance for task:", taskId);
    console.log("Current importance state:", currentState);
    console.log("New importance state:", newImportance);
    
    // Send the update to the server
    updateTaskImportance(taskId, newImportance)
        .then((success) => {
            if (success) {
                // Update the UI to reflect the new importance state
                if (newImportance) {
                    taskCard.classList.add("important");
                } else {
                    taskCard.classList.remove("important");
                }
                
                // Update the data attribute with the new state
                taskCard.setAttribute('data-importance', newImportance.toString());
                
                // Show success message
                if (newImportance) {
                    showToast("Task marked as important", "success");
                } else {
                    showToast("Task importance removed", "success");
                }
            } else {
                showToast("Failed to update task importance. Please try again.", "error");
            }
        })
        .catch(error => {
            console.error("Error updating task importance:", error);
            showToast("Error: " + error.message, "error");
        })
        .finally(() => {
            // Hide loading overlay
            document.getElementById("loadingOverlay").style.display = "none";
        });
}

// Function to update task importance on the server
async function updateTaskImportance(taskId, isImportant) {
    try {
        // Use URL encoded form data
        const params = new URLSearchParams();
        params.append("listId", listId);
        params.append("taskId", taskId);
        params.append("isImportant", isImportant.toString());
        
        // Debug logs
        console.log("Sending importance update request with:");
        console.log("listId:", listId);
        console.log("taskId:", taskId);
        console.log("isImportant:", isImportant.toString());
        
        const response = await fetch("/api/toggleImportance", {
            method: "POST",
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: params.toString(),
            credentials: "same-origin"
        });
        
        if (!response.ok) {
            // Get the error text from the response
            const errorText = await response.text();
            console.error("Server error:", errorText);
            throw new Error(errorText || "Server error");
        }
        
        return response.ok;
    } catch (error) {
        console.error("Error updating task importance:", error);
        return false;
    }
}

// ==================== Task Detail Modal Functions ====================

// Get modal elements
const modal = document.getElementById('taskDetailModal');
const closeBtn = document.querySelector('.close');
const viewModeContainer = document.getElementById('viewModeContainer');
const editModeContainer = document.getElementById('editModeContainer');
const editTaskButton = document.getElementById('editTaskButton');
const saveTaskButton = document.getElementById('saveTaskButton');
const cancelEditButton = document.getElementById('cancelEditButton');

// Current task being viewed/edited
let currentTaskId = null;

// Open task details modal
function openTaskDetails(event, taskElement) {
    // Prevent event from propagating (avoid starting drag)
    event.stopPropagation();
    
    // Get task ID
    const taskId = taskElement.getAttribute('data-task-id');
    currentTaskId = taskId;
    
    // Show loading overlay
    document.getElementById("loadingOverlay").style.display = "flex";
    
    // Fetch task details from server
    fetchTaskDetails(taskId)
        .then(task => {
            if (task) {
                // Update modal with task details
                updateModalWithTaskDetails(task);
                
                // Show the modal
                modal.style.display = "block";
            } else {
                showToast("Error loading task details", "error");
            }
        })
        .catch(error => {
            console.error("Error fetching task details:", error);
            showToast("Error: " + error.message, "error");
        })
        .finally(() => {
            // Hide loading overlay
            document.getElementById("loadingOverlay").style.display = "none";
        });
}

// Fetch task details from the server
async function fetchTaskDetails(taskId) {
    try {
        const params = new URLSearchParams();
        params.append("listId", listId);
        params.append("taskId", taskId);
        
        const response = await fetch(`/api/getTaskDetails?${params.toString()}`, {
            method: "GET",
            credentials: "same-origin"
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error("Server error:", errorText);
            throw new Error(errorText || "Server error");
        }
        
        return await response.json();
    } catch (error) {
        console.error("Error fetching task details:", error);
        throw error;
    }
}

// Update modal with task details
function updateModalWithTaskDetails(task) {
    // Update view mode
    document.getElementById('modalTaskTitle').textContent = task.title;
    document.getElementById('viewTaskTitle').textContent = task.title;
    document.getElementById('viewTaskStatus').textContent = task.status || 'Not Started';
    document.getElementById('viewTaskImportance').textContent = task.importance === 'high' ? 'High' : 'Normal';
    document.getElementById('viewTaskDueDate').textContent = task.dueDateTime || 'None';
    document.getElementById('viewTaskCategories').textContent = task.categories && task.categories.length > 0 
        ? task.categories.join(', ') 
        : 'None';
    
    // Update edit mode form
    document.getElementById('editTaskTitle').value = task.title;
    
    const statusSelect = document.getElementById('editTaskStatus');
    if (task.status === 'completed') {
        statusSelect.value = 'completed';
    } else if (task.categories && task.categories.includes('Doing')) {
        statusSelect.value = 'inProgress';
    } else {
        statusSelect.value = 'notStarted';
    }
    
    document.getElementById('editTaskImportance').value = task.importance === 'high' ? 'high' : 'normal';
    
    // Handle due date (if provided in ISO format, convert to YYYY-MM-DD for input)
    const dueDateInput = document.getElementById('editTaskDueDate');
    if (task.dueDateTimeRaw) {
        // Extract YYYY-MM-DD from ISO string
        const dateOnly = task.dueDateTimeRaw.split('T')[0];
        dueDateInput.value = dateOnly;
    } else {
        dueDateInput.value = '';
    }
    
    document.getElementById('editTaskCategories').value = task.categories ? task.categories.join(', ') : '';
    
    // Show view mode container, hide edit mode container
    viewModeContainer.style.display = 'block';
    editModeContainer.style.display = 'none';
}

// Switch to edit mode
function switchToEditMode() {
    viewModeContainer.style.display = 'none';
    editModeContainer.style.display = 'block';
}

// Switch to view mode
function switchToViewMode() {
    viewModeContainer.style.display = 'block';
    editModeContainer.style.display = 'none';
}

// Save task changes
async function saveTaskChanges() {
    // Get values from form
    const title = document.getElementById('editTaskTitle').value.trim();
    const status = document.getElementById('editTaskStatus').value;
    const importance = document.getElementById('editTaskImportance').value;
    const dueDate = document.getElementById('editTaskDueDate').value;
    const categoriesInput = document.getElementById('editTaskCategories').value.trim();
    
    // Validate input
    if (!title) {
        showToast("Title cannot be empty", "error");
        return;
    }
    
    // Parse categories
    let categories = [];
    if (categoriesInput) {
        categories = categoriesInput.split(',').map(cat => cat.trim()).filter(cat => cat);
    }
    
    // Prepare task data
    const taskData = {
        title,
        status,
        importance,
        dueDate,
        categories
    };
    
    // Show loading overlay
    document.getElementById("loadingOverlay").style.display = "flex";
    
    try {
        // Send update to server
        const success = await updateTask(currentTaskId, taskData);
        
        if (success) {
            // Refresh task details
            const updatedTask = await fetchTaskDetails(currentTaskId);
            updateModalWithTaskDetails(updatedTask);
            
            // Update task card in the UI
            updateTaskCardInUI(currentTaskId, updatedTask);
            
            // Switch back to view mode
            switchToViewMode();
            
            showToast("Task updated successfully", "success");
        } else {
            showToast("Failed to update task", "error");
        }
    } catch (error) {
        console.error("Error saving task changes:", error);
        showToast("Error: " + error.message, "error");
    } finally {
        // Hide loading overlay
        document.getElementById("loadingOverlay").style.display = "none";
    }
}

// Update task on the server
async function updateTask(taskId, taskData) {
    try {
        const params = new URLSearchParams();
        params.append("listId", listId);
        params.append("taskId", taskId);
        params.append("title", taskData.title);
        params.append("status", taskData.status);
        params.append("importance", taskData.importance);
        
        if (taskData.dueDate) {
            params.append("dueDate", taskData.dueDate);
        }
        
        if (taskData.categories && taskData.categories.length > 0) {
            params.append("categories", JSON.stringify(taskData.categories));
        }
        
        const response = await fetch("/api/updateTaskDetails", {
            method: "POST",
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: params.toString(),
            credentials: "same-origin"
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error("Server error:", errorText);
            throw new Error(errorText || "Server error");
        }
        
        return true;
    } catch (error) {
        console.error("Error updating task:", error);
        throw error;
    }
}

// Update task card in the UI
function updateTaskCardInUI(taskId, task) {
    const taskCard = document.querySelector(`[data-task-id="${taskId}"]`);
    if (!taskCard) return;
    
    // Update task title
    const titleElement = taskCard.querySelector('.task-title');
    if (titleElement) {
        titleElement.textContent = task.title;
    }
    
    // Update importance
    if (task.importance === 'high') {
        taskCard.classList.add('important');
        taskCard.setAttribute('data-importance', 'true');
    } else {
        taskCard.classList.remove('important');
        taskCard.setAttribute('data-importance', 'false');
    }
    
    // Update completed status
    if (task.status === 'completed') {
        taskCard.classList.add('completed');
    } else {
        taskCard.classList.remove('completed');
    }
    
    // Update categories
    const categoriesElement = taskCard.querySelector('.task-categories');
    if (categoriesElement) {
        // Clear existing categories
        categoriesElement.innerHTML = '';
        
        // Add new categories
        if (task.categories && task.categories.length > 0) {
            task.categories.forEach(category => {
                const categoryTag = document.createElement('span');
                categoryTag.className = 'category-tag';
                categoryTag.textContent = category;
                categoriesElement.appendChild(categoryTag);
            });
        }
    } else if (task.categories && task.categories.length > 0) {
        // Create categories container if it doesn't exist
        const newCategoriesElement = document.createElement('div');
        newCategoriesElement.className = 'task-categories';
        
        // Add categories
        task.categories.forEach(category => {
            const categoryTag = document.createElement('span');
            categoryTag.className = 'category-tag';
            categoryTag.textContent = category;
            newCategoriesElement.appendChild(categoryTag);
        });
        
        taskCard.appendChild(newCategoriesElement);
    }
    
    // Update due date
    let dueDateElement = taskCard.querySelector('.task-due');
    if (task.dueDateTime) {
        if (dueDateElement) {
            dueDateElement.textContent = `Due: ${task.dueDateTime}`;
        } else {
            dueDateElement = document.createElement('span');
            dueDateElement.className = 'task-due';
            dueDateElement.textContent = `Due: ${task.dueDateTime}`;
            taskCard.appendChild(dueDateElement);
        }
    } else if (dueDateElement) {
        dueDateElement.remove();
    }
    
    // If task status changed, move it to the appropriate column
    const currentColumn = taskCard.closest('.kanban-column');
    if (currentColumn) {
        const currentColumnName = currentColumn.getAttribute('data-column');
        let targetColumnName = '';
        
        if (task.status === 'completed') {
            targetColumnName = 'Done';
        } else if (task.categories && task.categories.includes('Doing')) {
            targetColumnName = 'Doing';
        } else {
            targetColumnName = 'Not Started';
        }
        
        if (currentColumnName !== targetColumnName) {
            // Find the target column
            const targetColumn = document.querySelector(`.kanban-column[data-column="${targetColumnName}"] .column-content`);
            if (targetColumn) {
                // Check for and remove the "no-tasks" message if it exists
                const noTasksMessage = targetColumn.querySelector(".no-tasks");
                if (noTasksMessage) {
                    noTasksMessage.remove();
                }
                
                // Move the task card to the new column
                targetColumn.appendChild(taskCard);
            }
        }
    }
}

// ==================== Add New Task Function ====================

// Add a new task
async function addNewTask() {
    // Get the task title from the input
    const input = document.getElementById('newTaskTitle');
    const title = input.value.trim();
    
    // Validate input
    if (!title) {
        showToast("Please enter a task title", "error");
        return;
    }
    
    // Show loading overlay
    document.getElementById("loadingOverlay").style.display = "flex";
    
    try {
        // Create the task on the server
        const taskId = await createNewTask(title);
        
        if (taskId) {
            // Clear the input
            input.value = '';
            
            // Refresh the column or add the new task to the UI
            const columnContent = document.querySelector('.kanban-column[data-column="Not Started"] .column-content');
            
            // Remove the "no tasks" message if it exists
            const noTasksMessage = columnContent.querySelector(".no-tasks");
            if (noTasksMessage) {
                noTasksMessage.remove();
            }
            
            // Create a new task card element
            const taskCard = document.createElement('div');
            taskCard.className = 'task-card';
            taskCard.draggable = true;
            taskCard.setAttribute('ondragstart', 'drag(event)');
            taskCard.setAttribute('data-task-id', taskId);
            taskCard.setAttribute('data-importance', 'false');
            taskCard.setAttribute('onclick', 'openTaskDetails(event, this)');
            
            // Create task title element
            const taskTitle = document.createElement('div');
            taskTitle.className = 'task-title';
            taskTitle.textContent = title;
            
            // Create importance star element
            const importanceStar = document.createElement('div');
            importanceStar.className = 'importance-star';
            importanceStar.textContent = '★';
            importanceStar.setAttribute('onclick', 'toggleImportance(event, this.parentElement)');
            
            // Add elements to task card
            taskCard.appendChild(taskTitle);
            taskCard.appendChild(importanceStar);
            
            // Add the task card to the column
            columnContent.appendChild(taskCard);
            
            showToast("Task added successfully", "success");
        } else {
            showToast("Failed to add task", "error");
        }
    } catch (error) {
        console.error("Error adding task:", error);
        showToast("Error: " + error.message, "error");
    } finally {
        // Hide loading overlay
        document.getElementById("loadingOverlay").style.display = "none";
    }
}

// Create a new task on the server
async function createNewTask(title) {
    try {
        const params = new URLSearchParams();
        params.append("listId", listId);
        params.append("title", title);
        
        const response = await fetch("/api/createTask", {
            method: "POST",
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: params.toString(),
            credentials: "same-origin"
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error("Server error:", errorText);
            throw new Error(errorText || "Server error");
        }
        
        const result = await response.json();
        return result.taskId;
    } catch (error) {
        console.error("Error creating task:", error);
        throw error;
    }
}

// Event listener for Enter key in the new task input
document.addEventListener('DOMContentLoaded', function() {
    const input = document.getElementById('newTaskTitle');
    if (input) {
        input.addEventListener('keyup', function(event) {
            if (event.key === 'Enter') {
                addNewTask();
            }
        });
    }
});

// Event Listeners for Modal
if (closeBtn) {
    closeBtn.addEventListener('click', () => {
        modal.style.display = "none";
    });
}

if (editTaskButton) {
    editTaskButton.addEventListener('click', switchToEditMode);
}

if (saveTaskButton) {
    saveTaskButton.addEventListener('click', saveTaskChanges);
}

if (cancelEditButton) {
    cancelEditButton.addEventListener('click', switchToViewMode);
}

// Close modal when clicking outside of it
window.addEventListener('click', (event) => {
    if (event.target === modal) {
        modal.style.display = "none";
    }
});
