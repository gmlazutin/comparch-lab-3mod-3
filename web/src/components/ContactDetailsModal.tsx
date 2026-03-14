import React, { useEffect, useState } from "react";
import { Modal, Button, Spinner } from "react-bootstrap";
import { apiCall } from "../utils/apiCall";
import { DefaultService } from "../api";
import { type PhoneDetails } from "./types";

interface ContactDetails {
    name: string;
    birthday: string;
    note?: string;
    phones: PhoneDetails[];
}

interface Props {
    contactId: number;
    onClose: () => void;
}

const ContactDetailsModal: React.FC<Props> = ({ contactId, onClose }) => {
    const [contact, setContact] = useState<ContactDetails | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchContact = async () => {
            setLoading(true);
            try {
                const cont = await apiCall(() => 
                    DefaultService.getContact(contactId, {
                        preload: {
                            enabled: true
                        },
                        withNote: true
                    })
                );
                if (!cont.ok) {
                    onClose();
                    return
                }

                setContact({
                    name: cont.data.contact.name,
                    birthday: cont.data.contact.birthday,
                    note: cont.data.contact.note,
                    phones: cont.data.contact.phones?.map(c => ({
                        number: c.phone,
                        primary: c.isPrimary || false
                    })) ?? []
                });
            } finally {
                setLoading(false);
            }
        };

        fetchContact();
    }, [contactId, onClose]);

    if (loading) {
        return (
            <Modal show onHide={onClose}>
                <Modal.Body className="text-center">
                    <Spinner animation="border" />
                </Modal.Body>
            </Modal>
        );
    }

    if (!contact) {
        return null;
    }

    const sortedPhones = [...contact.phones].sort((a, b) => (b.primary ? 1 : 0) - (a.primary ? 1 : 0));

    return (
        <Modal show onHide={onClose}>
            <Modal.Header closeButton>
                <Modal.Title>{contact.name}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <p><strong>Birthday:</strong> {contact.birthday}</p>
                {contact.note && <p><strong>Note:</strong> {contact.note}</p>}
                <p><strong>Phones:</strong></p>
                <ul>
                    {sortedPhones.map((p, i) => (
                        <li key={i}>{p.number} {p.primary && "(Primary)"}</li>
                    ))}
                </ul>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={onClose}>Close</Button>
            </Modal.Footer>
        </Modal>
    );
};

export default ContactDetailsModal;