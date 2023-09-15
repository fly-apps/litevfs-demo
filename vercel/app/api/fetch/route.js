import { NextResponse } from 'next/server';
import Database from 'litevfs';

export const dynamic = 'force-dynamic';
export const revalidate = 0;
export const runtime = 'nodejs';

const db = new Database('demo.db');
const stmt = db.prepare('SELECT id, data as value FROM (SELECT * FROM data ORDER BY id DESC LIMIT 20) ORDER BY id ASC');

export function GET(request) {
  const start = performance.now();
  const records = stmt.all();
  const latency = performance.now() - start;

  const response = NextResponse.json(
    {
      latency: latency + 'ms',
      records: records,
    },
    {
      status: 200,
    },
  );

  response.headers.set('Cache-Control', 'public, max-age=0, must-revalidate');
  response.headers.set('CDN-Cache-Control', 'public, max-age=0, must-revalidate');
  response.headers.set('Vercel-CDN-Cache-Control', 'public, max-age=0, must-revalidate');

  return response;
}
