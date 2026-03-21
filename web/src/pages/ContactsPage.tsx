import React, { useEffect, useState } from "react";
import { useAuth } from "../auth/AuthContext";
import ContactDetailsModal from "../components/ContactDetailsModal";
import AddContactModal from "../components/AddContactModal";
import { DefaultService } from "../api";
import { apiCall } from "../utils/apiCall";

interface TableContact {
    id: number;
    name: string;
    birthday: string;
    primaryPhone: string;
}

const ContactsPage: React.FC = () => {
    const { setSession } = useAuth();

    const [contacts, setContacts] = useState<TableContact[]>([]);
    const [selectedContactId, setSelectedContactId] = useState<number | null>(null);
    const [showAddModal, setShowAddModal] = useState(false);
    const [contactsEnd, setContactsEnd] = useState(true);

    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 5;

    const loadPage = async (page: number) => {
        const offset = (page - 1) * itemsPerPage;
        const contacts = await apiCall(() => 
            DefaultService.getContacts({
                selector: {
                    offset: offset,
                    limit: itemsPerPage+1
                }
            })
        );
        if (!contacts.ok) {
            return;
        }
        if (contacts.data.contacts.length > itemsPerPage) {
            setContactsEnd(false);
            contacts.data.contacts = contacts.data.contacts.slice(0,-1);
        }else{
            setContactsEnd(true);
        }
        setContacts(contacts.data.contacts.map(c => ({
            id: c.id,
            name: c.name,
            birthday: c.birthday,
            primaryPhone: c.phones?.[0]?.phone ?? "UNKNOWN"
        })));
    };

    useEffect(() => { loadPage(currentPage); }, [currentPage]);

    const handleDelete = async (id: number) => {
        const res = await apiCall(() =>
            DefaultService.deleteContact(id)
        );
        if (!res.ok) {
            return;
        }
        loadPage(currentPage);
    };

    const handleLogout = () => {
        setSession(null);
    };

    return (
        <div className="container mt-5">
            <div className="d-flex justify-content-between align-items-center mb-3">
                <h2>Contacts</h2>
                <div>
                    <button className="btn btn-success me-2" onClick={() => setShowAddModal(true)}>Add Contact</button>
                    <button className="btn btn-secondary" onClick={handleLogout}>Logout</button>
                </div>
            </div>

            <table className="table table-hover">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Birthday</th>
                        <th>Primary Phone</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {contacts.map(c => (
                        <tr key={c.id} onClick={() => setSelectedContactId(c.id)} style={{ cursor: "pointer" }}>
                            <td>{c.name}</td>
                            <td>{c.birthday}</td>
                            <td>{c.primaryPhone}</td>
                            <td>
                                <button className="btn btn-sm btn-danger" onClick={e => { e.stopPropagation(); handleDelete(c.id); }}>Delete</button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            <div className="d-flex justify-content-between align-items-center">
                <button className="btn btn-outline-primary" disabled={currentPage === 1} onClick={() => setCurrentPage(p => p - 1)}>Previous</button>
                <button className="btn btn-outline-primary" disabled={contactsEnd} onClick={() => setCurrentPage(p => p + 1)}>Next</button>
            </div>

            {selectedContactId && <ContactDetailsModal contactId={selectedContactId} onClose={() => { setSelectedContactId(null); loadPage(currentPage); }} />}
            {showAddModal && <AddContactModal onClose={() => { setShowAddModal(false); loadPage(currentPage); }} />}
        </div>
    );
};

export default ContactsPage;