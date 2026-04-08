import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '15s', target: 50 },
        { duration: '30s', target: 100 },
        { duration: '15s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<200'],
        http_req_failed: ['rate<0.001'],
    },
};

const BASE_URL = 'http://localhost:8080';

export function setup() {
    const adminRes = http.post(`${BASE_URL}/dummyLogin`,
        JSON.stringify({ role: 'admin' }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    const adminToken = JSON.parse(adminRes.body).token;
    const adminHeaders = { 'Content-Type': 'application/json', Authorization: `Bearer ${adminToken}` };

    const roomRes = http.post(`${BASE_URL}/rooms/create`,
        JSON.stringify({ name: 'Load Test Room', capacity: 10 }),
        { headers: adminHeaders }
    );
    console.log('room status:', roomRes.status, 'body:', roomRes.body);
    const roomId = JSON.parse(roomRes.body).room.id;

    const schedRes = http.post(`${BASE_URL}/rooms/${roomId}/schedule/create`,
        JSON.stringify({
            roomId: roomId,
            daysOfWeek: [1, 2, 3, 4, 5],
            startTime: '09:00',
            endTime: '18:00',
        }),
        { headers: adminHeaders }
    );
    console.log('schedule status:', schedRes.status, 'body:', schedRes.body);

    const userRes = http.post(`${BASE_URL}/dummyLogin`,
        JSON.stringify({ role: 'user' }),
        { headers: { 'Content-Type': 'application/json' } }
    );
    const userToken = JSON.parse(userRes.body).token;

    const now = new Date();
    let target = new Date(now);
    while (target.getDay() === 0 || target.getDay() === 6) {
        target.setDate(target.getDate() + 1);
    }
    const date = target.toISOString().split('T')[0];

    console.log('roomId:', roomId, 'date:', date);
    return { token: userToken, roomId, date };
}

export default function (data) {
    const res = http.get(
        `${BASE_URL}/rooms/${data.roomId}/slots/list?date=${data.date}`,
        { headers: { Authorization: `Bearer ${data.token}` } }
    );

    check(res, {
        'status 200': (r) => r.status === 200,
        'response < 200ms': (r) => r.timings.duration < 200,
    });

    sleep(0.1);
}