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
