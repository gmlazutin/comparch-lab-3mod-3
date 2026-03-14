type Result<T> =
  | { ok: true; data: T }
  | { ok: false; error: unknown };

export async function apiCall<T>(fn: () => Promise<T>): Promise<Result<T>> {
    try {
        return { ok: true, data: await fn() };
    } catch (e : any) {
        alert(e.body?.error ?? e);
        return { ok: false, error: e };
    }
}