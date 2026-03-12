import Modal from "react-modal";

export default function ContactModal({ contact, close }: any) {
  if (!contact) return null;

  const phones = [...contact.phones].sort((a, b) =>
    a.isPrimary ? -1 : 1
  );

  return (
    <Modal isOpen onRequestClose={close}>
      <h2>{contact.name}</h2>

      <p>Birthday: {contact.birthday}</p>

      {contact.note && <p>Note: {contact.note}</p>}

      <ul>
        {phones.map((p: any) => (
          <li key={p.id}>
            {p.phone} {p.isPrimary ? "(primary)" : ""}
          </li>
        ))}
      </ul>

      <button onClick={close}>close</button>
    </Modal>
  );
}