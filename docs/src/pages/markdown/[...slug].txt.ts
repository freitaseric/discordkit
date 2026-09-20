import { getCollection } from 'astro:content';
import type { APIRoute } from 'astro';
export async function getStaticPaths() {
  const entries = await getCollection('docs');
  return entries.map((entry) => ({ params: { slug: entry.id || "index" }, props: { entry } }));
}
export const GET: APIRoute = ({ props }) => {
  const { entry } = props;
  const body = (entry.body ?? '').replace(/^import .*;?\s*$/gm, '');
  return new Response(`# ${entry.data.title}\n\n${entry.data.description ?? ''}\n\n${body}`, {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
};
