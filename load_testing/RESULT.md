## Отчёт
█ THRESHOLDS

    error_rate
    ✓ 'rate<0.05' rate=0.00%

    http_req_duration
    ✓ 'p(95)<500' p(95)=10.43ms
    ✓ 'p(99)<1000' p(99)=13.37ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%

    http_reqs
    ✓ 'rate>100' rate=870.757903/s


█ TOTAL RESULTS

    checks_total.......: 694004  2280.628239/s
    checks_succeeded...: 100.00% 694004 out of 694004
    checks_failed......: 0.00%   0 out of 694004

    ✓ Create Team: status is 201
    ✓ Create Team: has valid response
    ✓ Team has correct name
    ✓ Team has members
    ✓ Get Team: status is 200
    ✓ Get Team: has valid response
    ✓ Get Team: team name matches
    ✓ Set User IsActive: status is 200
    ✓ Set User IsActive: has valid response
    ✓ User is_active changed
    ✓ Create PR: status is 201
    ✓ Create PR: has valid response
    ✓ PR has correct ID
    ✓ PR status is OPEN
    ✓ PR has assigned reviewers
    ✓ Get User Review: status is 200
    ✓ Get User Review: has valid response
    ✓ Response has user_id
    ✓ Response has pull_requests array
    ✓ Reassign: valid response
    ✓ Merge PR: status is 200
    ✓ Merge PR: has valid response
    ✓ PR status is MERGED
    ✓ PR has mergedAt timestamp
    ✓ Merge PR Idempotent: status is 200
    ✓ Merge PR Idempotent: has valid response
    ✓ Health Check: status is 200
    ✓ Health Check: has valid response

    CUSTOM
    error_rate.....................: 0.00%  0 out of 201896
    pr_creation_time...............: avg=8.539327 min=1     med=8      max=59     p(90)=10     p(95)=12     
    successful_requests............: 201896 663.468393/s
    team_creation_time.............: avg=8.242144 min=1     med=8      max=58     p(90)=10     p(95)=11     

    HTTP
    http_req_duration..............: avg=5.08ms   min=0s    med=4.72ms max=62.3ms p(90)=9.53ms p(95)=10.43ms
      { expected_response:true }...: avg=5.08ms   min=0s    med=4.72ms max=62.3ms p(90)=9.53ms p(95)=10.43ms
    http_req_failed................: 0.00%  0 out of 264975
    http_reqs......................: 264975 870.757903/s

    EXECUTION
    dropped_iterations.............: 56846  186.806694/s
    iteration_duration.............: avg=5.15s    min=5.13s med=5.15s  max=5.23s  p(90)=5.16s  p(95)=5.17s  
    iterations.....................: 25237  82.933549/s
    vus............................: 1      min=1           max=762
    vus_max........................: 774    min=199         max=774

    NETWORK
    data_received..................: 107 MB 350 kB/s
    data_sent......................: 64 MB  209 kB/s




running (5m04.3s), 000/774 VUs, 25237 complete and 0 interrupted iterations

constant_rps ✓ [======================================] 000/200 VUs  2m0s   100.00 iters/s

default      ✓ [======================================] 000/100 VUs  5m0s 

ramping_rps  ✓ [======================================] 000/500 VUs  3m30s  0004.08 iters/s

## Интерпретация результатов
1. Все показатели по времени отклика сервиса в норме
2. Ошибок нет
3. Среднее время создания команды ~7ms, создание PR ~6ms
4. Сервис стабильно работает при нагрузке в 100 виртуальных пользователей
5. В целом, сервис показывает хорошие результаты при нагрузочном тестировании (ну ещё бы, go + fiber + postgres)

Правда, поскольку у меня нет ни трейсов, ни каких-либо метрик изнутри сервиса, сложно сказать, где именно могут быть узкие места. Но в целом, результаты хорошие.
В идеале ещё смотреть не только на Summary, но и на распределения, да и графики в целом