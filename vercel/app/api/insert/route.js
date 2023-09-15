import { NextResponse } from 'next/server';
import Database from 'litevfs';

export const dynamic = 'force-dynamic';
export const revalidate = 0;
export const runtime = 'nodejs';

const db = new Database('demo.db');
const stmt = db.prepare('INSERT INTO data (data) VALUES (?)');

export function POST(request) {
  const start = performance.now();
  db.with_write_lease(function() {
    stmt.run(Math.floor(Math.random() * 100000))
  });
  const latency = performance.now() - start;
  return NextResponse.json(
    {
      latency: latency + 'ms',
    },
    {
      status: 200,
    },
  );
}
