import http from 'k6/http';
import {check, group, sleep} from 'k6';
import {Counter, Rate, Trend} from 'k6/metrics';

// Метрики для мониторинга
const errorRate = new Rate('error_rate');
const teamCreationTime = new Trend('team_creation_time');
const prCreationTime = new Trend('pr_creation_time');
const successfulRequests = new Counter('successful_requests');

// Конфигурация нагрузочного тестирования
export const options = {
    scenarios: {
        // Сценарий с постоянным RPS
        constant_rps: {
            executor: 'constant-arrival-rate',
            rate: 100,              // 100 запросов в секунду
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 50,
            maxVUs: 200,
        },
        // Сценарий с увеличивающимся RPS
        ramping_rps: {
            executor: 'ramping-arrival-rate',
            startRate: 10,
            timeUnit: '1s',
            preAllocatedVUs: 50,
            maxVUs: 500,
            stages: [
                {duration: '30s', target: 50},   // 50 RPS
                {duration: '1m', target: 200},   // 200 RPS
                {duration: '1m', target: 500},   // 500 RPS
                {duration: '30s', target: 1000}, // 1000 RPS (пик)
                {duration: '30s', target: 0},
            ],
        },
        default: {
            executor: 'ramping-vus',
            stages: [
                {duration: '30s', target: 10},   // Разогрев: 10 пользователей за 30 секунд
                {duration: '1m', target: 50},    // Увеличение нагрузки: 50 пользователей за 1 минуту
                {duration: '2m', target: 100},   // Увеличиваем пик: 100 пользователей за 2 минуты
                {duration: '1m', target: 50},    // Снижение нагрузки: 50 пользователей
                {duration: '30s', target: 0},    // Остановка
            ]
        }
    },
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.05'],
        http_reqs: ['rate>100'],  // Минимум 100 RPS
        error_rate: ['rate<0.05']
    },
};
;

// Базовый URL API (можно переопределить через переменную окружения)
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Генерация уникальных ID
const VU_ID = __VU;
const ITER_ID = () => __ITER;

function generateTeamName() {
    return `team_${VU_ID}_${ITER_ID()}_${Date.now()}`;
}

function generateUserId() {
    return `user_${VU_ID}_${ITER_ID()}_${Date.now()}_${Math.random().toString(36).substring(7)}`;
}

function generatePRId() {
    return `pr_${VU_ID}_${ITER_ID()}_${Date.now()}_${Math.random().toString(36).substring(7)}`;
}

// Вспомогательная функция для проверки ответов
function checkResponse(response, expectedStatus, metricName) {
    const success = check(response, {
        [`${metricName}: status is ${expectedStatus}`]: (r) => r.status === expectedStatus,
        [`${metricName}: has valid response`]: (r) => r.body && r.body.length > 0,
    });

    errorRate.add(!success);
    if (success) {
        successfulRequests.add(1);
    }

    return success;
}

export default function () {
    const headers = {'Content-Type': 'application/json'};

    // Генерируем данные для тестирования
    const teamName = generateTeamName();
    const userId1 = generateUserId();
    const userId2 = generateUserId();
    const userId3 = generateUserId();
    const userId4 = generateUserId();
    const userId5 = generateUserId();
    let oneTimePRId = generatePRId();

    group('Team Operations', () => {
        // Тест 1: Создание команды
        group('Create Team', () => {
            const teamPayload = JSON.stringify({
                team_name: teamName,
                members: [
                    {user_id: userId1, username: `Alice_${userId1}`, is_active: true},
                    {user_id: userId2, username: `Bob_${userId2}`, is_active: true},
                    {user_id: userId3, username: `Charlie_${userId3}`, is_active: true},
                    {user_id: userId4, username: `Tony_${userId4}`, is_active: true},
                    {user_id: userId5, username: `Tomm_${userId5}`, is_active: true},
                ],
            });

            const createTeamStart = Date.now();
            const createTeamRes = http.post(`${BASE_URL}/team/add`, teamPayload, {headers});
            teamCreationTime.add(Date.now() - createTeamStart);

            const success = checkResponse(createTeamRes, 201, 'Create Team');

            if (success) {
                check(createTeamRes, {
                    'Team has correct name': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.team && body.team.team_name === teamName;
                        } catch (e) {
                            return false;
                        }
                    },
                    'Team has members': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.team && body.team.members && body.team.members.length === 5;
                        } catch (e) {
                            return false;
                        }
                    },
                });
            }
        });

        sleep(0.5);

        // Тест 2: Получение команды
        group('Get Team', () => {
            const getTeamRes = http.get(`${BASE_URL}/team/get?team_name=${teamName}`, {headers});
            checkResponse(getTeamRes, 200, 'Get Team');

            if (getTeamRes.status === 200) {
                check(getTeamRes, {
                    'Get Team: team name matches': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.team_name === teamName;
                        } catch (e) {
                            return false;
                        }
                    },
                });
            }
        });
    });

    sleep(0.5);

    group('User Operations', () => {
        // Тест 3: Изменение статуса активности пользователя
        group('Set User IsActive', () => {
            const setActivePayload = JSON.stringify({
                user_id: userId2,
                is_active: false,
            });

            const setActiveRes = http.post(`${BASE_URL}/users/setIsActive`, setActivePayload, {headers});
            checkResponse(setActiveRes, 200, 'Set User IsActive');

            if (setActiveRes.status === 200) {
                check(setActiveRes, {
                    'User is_active changed': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.user && body.user.is_active === false;
                        } catch (e) {
                            return false;
                        }
                    },
                });
            }
        });

        sleep(0.3);

        // Возвращаем пользователя в активное состояние
        const reactivatePayload = JSON.stringify({
            user_id: userId2,
            is_active: true,
        });
        http.post(`${BASE_URL}/users/setIsActive`, reactivatePayload, {headers});
    });

    sleep(0.5);

    group('Pull Request Operations', () => {
        // Тест 4: Создание Pull Request
        group('Create Pull Request', () => {
            const prId = generatePRId();
            oneTimePRId = prId; // Сохраняем для последующих тестов
            const prPayload = JSON.stringify({
                pull_request_id: prId,
                pull_request_name: `Feature: Load Test PR ${prId}`,
                author_id: userId1,
            });

            const createPRStart = Date.now();
            const createPRRes = http.post(`${BASE_URL}/pullRequest/create`, prPayload, {headers});
            prCreationTime.add(Date.now() - createPRStart);

            const success = checkResponse(createPRRes, 201, 'Create PR');

            if (success) {
                check(createPRRes, {
                    'PR has correct ID': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.pr && body.pr.pull_request_id === prId;
                        } catch (e) {
                            return false;
                        }
                    },
                    'PR status is OPEN': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.pr && body.pr.status === 'OPEN';
                        } catch (e) {
                            return false;
                        }
                    },
                    'PR has assigned reviewers': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.pr && body.pr.assigned_reviewers && body.pr.assigned_reviewers.length > 0;
                        } catch (e) {
                            return false;
                        }
                    },
                });
            }
        });

        sleep(0.5);

        // Тест 5: Получение PR для ревьюера
        group('Get User Review', () => {
            const getReviewRes = http.get(`${BASE_URL}/users/getReview?user_id=${userId2}`, {headers});
            checkResponse(getReviewRes, 200, 'Get User Review');

            if (getReviewRes.status === 200) {
                check(getReviewRes, {
                    'Response has user_id': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.user_id === userId2;
                        } catch (e) {
                            return false;
                        }
                    },
                    'Response has pull_requests array': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return Array.isArray(body.pull_requests);
                        } catch (e) {
                            return false;
                        }
                    },
                });
            }
        });

        sleep(0.5);

        // Тест 6: Переназначение ревьюера (если был назначен)
        group('Reassign Reviewer', () => {
            const prId = oneTimePRId;
            // Сначала проверяем кто был назначен
            const getReviewRes = http.get(`${BASE_URL}/users/getReview?user_id=${userId2}`, {headers});

            if (getReviewRes.status === 200) {
                try {
                    const body = JSON.parse(getReviewRes.body);
                    const assignedPR = body.pull_requests.find(pr => pr.pull_request_id === prId);

                    if (assignedPR) {
                        const reassignPayload = JSON.stringify({
                            pull_request_id: prId,
                            old_user_id: userId2,
                        });

                        const reassignRes = http.post(`${BASE_URL}/pullRequest/reassign`, reassignPayload, {headers});

                        // Успешное переназначение или отсутствие кандидатов - оба варианта валидны
                        check(reassignRes, {
                            'Reassign: valid response': (r) => r.status === 200 || r.status === 409,
                        });
                    }
                } catch (e) {
                    // Игнорируем ошибки парсинга
                }
            }
        });

        sleep(0.5);

        // Тест 7: Слияние Pull Request
        group('Merge Pull Request', () => {
            const mergePayload = JSON.stringify({
                pull_request_id: oneTimePRId,
            });

            const mergeRes = http.post(`${BASE_URL}/pullRequest/merge`, mergePayload, {headers});
            checkResponse(mergeRes, 200, 'Merge PR');

            if (mergeRes.status === 200) {
                check(mergeRes, {
                    'PR status is MERGED': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.pr && body.pr.status === 'MERGED';
                        } catch (e) {
                            return false;
                        }
                    },
                    'PR has mergedAt timestamp': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return body.pr && body.pr.merged_at;
                        } catch (e) {
                            return false;
                        }
                    },
                });
            }
        });

        sleep(0.3);

        // Тест 8: Повторное слияние (идемпотентность)
        group('Merge PR Again (Idempotency)', () => {
            const mergePayload = JSON.stringify({
                pull_request_id: oneTimePRId,
            });

            const mergeAgainRes = http.post(`${BASE_URL}/pullRequest/merge`, mergePayload, {headers});
            checkResponse(mergeAgainRes, 200, 'Merge PR Idempotent');
        });
    });

    sleep(0.5);

    // Тест 9: Health Check
    group('Health Check', () => {
        const healthRes = http.get(`${BASE_URL}/health`);
        checkResponse(healthRes, 200, 'Health Check');
    });

    sleep(1); // Пауза между итерациями
}

// Функция для запуска только smoke-теста (быстрая проверка)
export function smokeTest() {
    const headers = {'Content-Type': 'application/json'};

    const teamName = `smoke_team_${Date.now()}`;
    const userId = `smoke_user_${Date.now()}`;

    const teamPayload = JSON.stringify({
        team_name: teamName,
        members: [
            {user_id: userId, username: 'SmokeUser', is_active: true},
        ],
    });

    const res = http.post(`${BASE_URL}/team/add`, teamPayload, {headers});
    check(res, {
        'Smoke test: API is available': (r) => r.status === 201,
    });
}

