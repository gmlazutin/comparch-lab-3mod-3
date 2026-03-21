import React, { useEffect, useState } from "react";
import { Modal, Button, Spinner, Form } from "react-bootstrap";
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

    const [name, setName] = useState("");
    const [birthday, setBirthday] = useState("");
    const [note, setNote] = useState("");

    useEffect(() => {
        const fetchContact = async () => {
            setLoading(true);
            try {
                const cont = await apiCall(() =>
                    DefaultService.getContact(contactId, {
                        preload: { enabled: true },
                        withNote: true
                    })
                );
                if (!cont.ok) {
                    onClose();
                    return;
                }

                const data = cont.data.contact;
                setContact({
                    name: data.name,
                    birthday: data.birthday,
                    note: data.note,
                    phones: data.phones?.map(c => ({
                        number: c.phone,
                        primary: c.isPrimary || false
                    })) ?? []
                });

                setName(data.name);
                setBirthday(data.birthday);
                setNote(data.note ?? "");
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

    if (!contact) return null;

    const sortedPhones = [...contact.phones].sort(
        (a, b) => (b.primary ? 1 : 0) - (a.primary ? 1 : 0)
    );

    const hasChanges =
        name !== contact.name ||
        birthday !== contact.birthday ||
        note !== (contact.note ?? "");

    const handleSubmit = async () => {
        if (!hasChanges) return;

        const res = await apiCall(() => 
            DefaultService.updateContact(contactId, {
                name: name !== contact.name ? name : undefined,
                birthday: birthday !== contact.birthday ? birthday : undefined,
                note: note !== (contact.note ?? "") ? note : undefined,
            })
        );
        if (!res.ok) {
            return;
        }
        
        onClose();
    };

    return (
        <Modal show onHide={onClose} size="lg">
            <Modal.Header closeButton>
                <Modal.Title>Contact</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group className="mb-3">
                        <Form.Label>Name</Form.Label>
                        <Form.Control
                            type="text"
                            value={name}
                            onChange={e => setName(e.target.value)}
                        />
                    </Form.Group>

                    <Form.Group className="mb-3">
                        <Form.Label>Birthday</Form.Label>
                        <Form.Control
                            type="date"
                            value={birthday}
                            onChange={e => setBirthday(e.target.value)}
                        />
                    </Form.Group>

                    <Form.Group className="mb-3">
                        <Form.Label>Note</Form.Label>
                        <Form.Control
                            as="textarea"
                            rows={3}
                            value={note}
                            onChange={e => setNote(e.target.value)}
                        />
                    </Form.Group>

                    <Form.Group>
                        <Form.Label>Phones:</Form.Label>
                        <ul>
                            {sortedPhones.map((p, i) => (
                                <li key={i}>
                                    {p.number} {p.primary && "(Primary)"}
                                </li>
                            ))}
                        </ul>
                    </Form.Group>
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={onClose}>
                    Cancel
                </Button>
                <Button
                    variant="primary"
                    onClick={handleSubmit}
                    disabled={!hasChanges}
                >
                    Save
                </Button>
            </Modal.Footer>
        </Modal>
    );
};

export default ContactDetailsModal;