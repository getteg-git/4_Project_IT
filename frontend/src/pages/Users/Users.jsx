import { useEffect, useState } from "react";
import api from "../../api/axios";
import "./Users.css";

function Users({ user, setUser }) {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

    const [keyword, setKeyword] = useState("");

    // Add form
    const [newUsername, setNewUsername] = useState("");
    const [newPassword, setNewPassword] = useState("");

    // Edit form
    const [editingId, setEditingId] = useState(null);
    const [editUsername, setEditUsername] = useState("");
    const [editPassword, setEditPassword] = useState("");

    useEffect(() => {
        fetchUsers();
    }, []);

    // Get all users
    const fetchUsers = async () => {
        try {
            setLoading(true);
            const res = await api.get("/users");
            setUsers(res.data);
            setError("");
        } catch (err) {
            console.error(err);
            setError("โหลดข้อมูลไม่สำเร็จ");
        } finally {
            setLoading(false);
        }
    };

    // Search
    const handleSearch = async () => {
        if (!keyword.trim()) {
            fetchUsers();
            return;
        }

        try {
            setLoading(true);
            const res = await api.get(`/users/search?q=${keyword}`);
            setUsers(res.data);
            setError("");
        } catch (err) {
            console.error(err);
            setError("ค้นหาไม่สำเร็จ");
        } finally {
            setLoading(false);
        }
    };

    // Add user
    const handleAddUser = async () => {
        if (!newUsername || !newPassword) {
            alert("กรุณากรอกข้อมูลให้ครบ");
            return;
        }

        try {
            await api.post("/users", {
                username: newUsername,
                password: newPassword,
            });

            setNewUsername("");
            setNewPassword("");

            fetchUsers();
        } catch (err) {
            console.error(err);
            alert("เพิ่ม user ไม่สำเร็จ");
        }
    };

    const handleDeleteUser = async (id) => {
        const ok = window.confirm("ต้องการลบ user นี้ใช่ไหม?");

        if (!ok) return;

        try {
            await api.delete(`/users/${id}`);
            fetchUsers();
        } catch (err) {
            console.error(err);
            alert("ลบไม่สำเร็จ");
        }
    };

    const handleEditClick = (u) => {
        setEditingId(u.id);
        setEditUsername(u.username);
        setEditPassword(u.password);
    };

    const handleUpdateUser = async () => {
        if (!editUsername || !editPassword) {
            alert("กรุณากรอกข้อมูลให้ครบ");
            return;
        }

        try {
            await api.put(`/users/${editingId}`, {
                username: editUsername,
                password: editPassword,
            });

            setEditingId(null);
            setEditUsername("");
            setEditPassword("");

            fetchUsers();
        } catch (err) {
            console.error(err);
            alert("แก้ไขไม่สำเร็จ");
        }
    };

    const handleCancelEdit = () => {
        setEditingId(null);
    };

    const handleLogout = () => {
        setUser(null);
    };

    return (
        <div className="users-container">
            <h2>Welcome, {user.username}</h2>

            <button className="logout-btn" onClick={handleLogout}>
                Logout
            </button>

            <h3>Users Management</h3>

            {/* Add User */}
            <div className="add-box">
                <input
                    type="text"
                    placeholder="Username"
                    value={newUsername}
                    onChange={(e) => setNewUsername(e.target.value)}
                />

                <input
                    type="password"
                    placeholder="Password"
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                />

                <button onClick={handleAddUser}>Add User</button>
            </div>

            {/* Edit User */}
            {editingId && (
                <div className="edit-box">
                    <h4>Edit User</h4>

                    <input
                        type="text"
                        value={editUsername}
                        onChange={(e) => setEditUsername(e.target.value)}
                        placeholder="Username"
                    />

                    <input
                        type="password"
                        value={editPassword}
                        onChange={(e) => setEditPassword(e.target.value)}
                        placeholder="Password"
                    />

                    <button onClick={handleUpdateUser}>Save</button>
                    <button onClick={handleCancelEdit}>Cancel</button>
                </div>
            )}

            {/* Search */}
            <div className="search-box">
                <input
                    type="text"
                    placeholder="Search username..."
                    value={keyword}
                    onChange={(e) => setKeyword(e.target.value)}
                />

                <button onClick={handleSearch}>Search</button>

                <button onClick={fetchUsers}>Reset</button>
            </div>

            {loading && <p>Loading...</p>}
            {error && <p className="error">{error}</p>}

            {!loading && !error && (
                <table className="users-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Username</th>
                            <th>Password</th>
                            <th>Action</th>
                        </tr>
                    </thead>

                    <tbody>
                        {users.length === 0 ? (
                            <tr>
                                <td colSpan="3">No data</td>
                            </tr>
                        ) : (
                            users.map((u) => (
                                <tr key={u.id}>
                                    <td>{u.id}</td>
                                    <td>{u.username}</td>
                                    <td>{u.password}</td>

                                    <td>
                                        <button
                                            className="edit-btn"
                                            onClick={() => handleEditClick(u)}
                                        >
                                            Edit
                                        </button>

                                        <button
                                            className="delete-btn"
                                            onClick={() => handleDeleteUser(u.id)}
                                        >
                                            Delete
                                        </button>
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            )}
        </div>
    );
}

export default Users;