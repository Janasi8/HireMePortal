console.log("✅ Dashboard JS loaded");

// Dummy globals to avoid errors if backend not connected yet
window.currentUsername = window.currentUsername || "Guest User";
window.currentUserId = window.currentUserId || 0;

// Safe init (agar function exist karta ho tabhi call kare)
document.addEventListener("DOMContentLoaded", () => {
    if (typeof initDashboard === "function") {
        initDashboard();
    }
});
