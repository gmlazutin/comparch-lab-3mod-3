import { useState } from "react";
import Modal from "react-modal";
import { DefaultService } from "../api";

export default function AddContactModal({ close, reload }: any) {
  const [name, setName] = useState("");
  const [birthday, setBirthday] = useState("");
  const [note, setNote] = useState("");

  const [phones, setPhones] = useState([
    { phone: "", isPrimary: true },
  ]);

  async function submit() {
    if (!phones.some((p) => p.isPrimary)) {
      alert("primary required");
      return;
    }

    try {
      await DefaultService.addContact({
        name,
        birthday,
        note,
        initialPhones: phones,
      });

      reload();
      close();
    } catch (e: any) {
      alert(e.body?.error);
    }
  }

  return (
    <Modal isOpen onRequestClose={close}>
      <h2>Add contact</h2>

      <input
        placeholder="name"
        onChange={(e) => setName(e.target.value)}
      />

      <input
        type="date"
        onChange={(e) => setBirthday(e.target.value)}
      />

      <input
        placeholder="note"
        onChange={(e) => setNote(e.target.value)}
      />

      {phones.map((p, i) => (
        <div key={i}>
          <input
            placeholder="phone"
            onChange={(e) => {
              const copy = [...phones];
              copy[i].phone = e.target.value;
              setPhones(copy);
            }}
          />

          <input
            type="radio"
            checked={p.isPrimary}
            onChange={() =>
              setPhones(
                phones.map((x, idx) => ({
                  ...x,
                  isPrimary: idx === i,
                }))
              )
            }
          />
        </div>
      ))}

      <button
        onClick={() =>
          setPhones([...phones, { phone: "", isPrimary: false }])
        }
      >
        add phone
      </button>

      <button onClick={submit}>save</button>
    </Modal>
  );
}