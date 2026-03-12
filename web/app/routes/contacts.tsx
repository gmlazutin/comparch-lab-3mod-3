import { useEffect, useState } from "react";
import { DefaultService } from "../api";
import { useAuth } from "../context/AuthContext";
import ContactModal from "../components/ContactModal";
import AddContactModal from "../components/AddContactModal";

export default function Contacts() {
  const { token } = useAuth();

  const [contacts, setContacts] = useState<any[]>([]);
  const [page, setPage] = useState(0);

  const [selected, setSelected] = useState<any>(null);
  const [addOpen, setAddOpen] = useState(false);

  async function load() {
    try {
      const res = await DefaultService.getContacts({
        selector: {
          offset: page * 10,
          limit: 10,
        },
      });

      setContacts(res.contacts);
    } catch (e: any) {
      alert(e.body?.error);
    }
  }

  useEffect(() => {
    if (!token) {
      window.location.href = "/login";
      return;
    }

    load();
  }, [page]);

  async function openContact(id: number) {
    try {
      const res = await DefaultService.getContact(
        id,
        {
          preload: { enabled: true, primaryOnly: false },
          withNote: true,
        }
      );

      setSelected(res.contact);
    } catch (e: any) {
      alert(e.body?.error);
    }
  }

  async function del(id: number) {
    if (!confirm("delete?")) return;

    try {
      await DefaultService.deleteContact(id);
      load();
    } catch (e: any) {
      alert(e.body?.error);
    }
  }

  return (
    <div>
      <h1>Contacts</h1>

      <button onClick={() => setAddOpen(true)}>add</button>

      <table border={1}>
        <thead>
          <tr>
            <th>Name</th>
            <th>Birthday</th>
            <th>Primary phone</th>
            <th></th>
          </tr>
        </thead>

        <tbody>
          {contacts.map((c) => (
            <tr key={c.id}>
              <td onClick={() => openContact(c.id)}>
                {c.name}
              </td>

              <td>{c.birthday}</td>

              <td>
                {c.phones.length
                  ? c.phones[0].phone
                  : "unknown"}
              </td>

              <td>
                <button onClick={() => del(c.id)}>
                  delete
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <button onClick={() => setPage((p) => (p >= 1) ? p - 1 : p)}>
        prev
      </button>

      <button onClick={() => setPage((p) => p + 1)}>
        next
      </button>

      {selected && (
        <ContactModal
          contact={selected}
          close={() => setSelected(null)}
        />
      )}

      {addOpen && (
        <AddContactModal
          close={() => setAddOpen(false)}
          reload={load}
        />
      )}
    </div>
  );
}