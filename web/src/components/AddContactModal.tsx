import React, { useState } from "react";
import { Modal, Button, Form } from "react-bootstrap";
import { apiCall } from "../utils/apiCall";
import { DefaultService } from "../api";
import {type PhoneDetails } from "./types";

interface Props {
    onClose: () => void;
}

const AddContactModal: React.FC<Props> = ({ onClose }) => {
    const [name, setName] = useState("");
    const [birthday, setBirthday] = useState("");
    const [note, setNote] = useState("");
    const [phones, setPhones] = useState<PhoneDetails[]>([{ number: "", primary: true }]);

    const addPhoneField = () => {
        setPhones([...phones, { number: "", primary: false }]);
    };

    const updatePhone = (index: number, number: string, primary: boolean) => {
        const newPhones = phones.map((p, i) => ({
            number: i === index ? number : p.number,
            primary: primary ? i === index : p.primary
        }));
        setPhones(newPhones);
    };

    const handleSubmit = async () => {
        const res = await apiCall(() => 
            DefaultService.addContact({
                name: name,
                birthday: birthday,
                note: note || undefined,
                initialPhones: phones.map(c => ({
                    phone: c.number,
                    isPrimary: c.primary && true || undefined
                }))
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
                <Modal.Title>Add Contact</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form>
                    <Form.Group className="mb-3">
                        <Form.Label>Name</Form.Label>
                        <Form.Control value={name} onChange={e => setName(e.target.value)} />
                    </Form.Group>
                    <Form.Group className="mb-3">
                        <Form.Label>Birthday</Form.Label>
                        <Form.Control type="date" value={birthday} onChange={e => setBirthday(e.target.value)} />
                    </Form.Group>
                    <Form.Group className="mb-3">
                        <Form.Label>Note (optional)</Form.Label>
                        <Form.Control as="textarea" value={note} onChange={e => setNote(e.target.value)} />
                    </Form.Group>
                    <Form.Label>Phones</Form.Label>
                    {phones.map((p, i) => (
                        <div key={i} className="d-flex mb-2 align-items-center">
                            <Form.Control
                                value={p.number}
                                onChange={e => updatePhone(i, e.target.value, p.primary)}
                                placeholder="Phone number"
                            />
                            <Form.Check
                                type="radio"
                                name="primaryPhone"
                                checked={p.primary}
                                onChange={() => updatePhone(i, p.number, true)}
                                label="Primary"
                                className="ms-2"
                            />
                        </div>
                    ))}
                    <Button variant="secondary" onClick={addPhoneField}>Add Phone</Button>
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={onClose}>Cancel</Button>
                <Button variant="primary" onClick={handleSubmit}>Save</Button>
            </Modal.Footer>
        </Modal>
    );
};

export default AddContactModal;