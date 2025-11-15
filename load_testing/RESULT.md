## Отчёт
█ THRESHOLDS

    error_rate
    ✓ 'rate<0.05' rate=0.00%

    http_req_duration
    ✓ 'p(95)<500' p(95)=8.35ms
    ✓ 'p(99)<1000' p(99)=10.85ms

    http_req_failed
    ✓ 'rate<0.05' rate=0.00%


█ TOTAL RESULTS

    checks_total.......: 87791   289.238573/s
    checks_succeeded...: 100.00% 87791 out of 87791
    checks_failed......: 0.00%   0 out of 87791

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
    error_rate.....................: 0.00%  0 out of 25536
    pr_creation_time...............: avg=6.276629 min=1     med=6      max=21     p(90)=9      p(95)=10    
    successful_requests............: 25536  84.131588/s
    team_creation_time.............: avg=7.227444 min=0     med=7      max=23     p(90)=11     p(95)=11    

    HTTP
    http_req_duration..............: avg=3.65ms   min=0s    med=3.24ms max=47.8ms p(90)=6.97ms p(95)=8.35ms
      { expected_response:true }...: avg=3.65ms   min=0s    med=3.24ms max=47.8ms p(90)=6.97ms p(95)=8.35ms
    http_req_failed................: 0.00%  0 out of 33527
    http_reqs......................: 33527  110.45895/s

    EXECUTION
    iteration_duration.............: avg=5.14s    min=5.12s med=5.14s  max=5.19s  p(90)=5.14s  p(95)=5.15s 
    iterations.....................: 3192   10.516448/s
    vus............................: 2      min=1          max=100
    vus_max........................: 100    min=100        max=100

    NETWORK
    data_received..................: 13 MB  44 kB/s
    data_sent......................: 8.0 MB 26 kB/s




running (5m03.5s), 000/100 VUs, 3192 complete and 0 interrupted iterations
default ✓ [======================================] 000/100 VUs  5m0s

## Интерпретация результатов
1. Все показатели по времени отклика сервиса в норме
2. Ошибок нет
3. Среднее время создания команды ~7ms, создание PR ~6ms
4. Сервис стабильно работает при нагрузке в 100 виртуальных пользователей
5. В целом, сервис показывает хорошие результаты при нагрузочном тестировании (ну ещё бы, go + fiber + postgres)

Правда, поскольку у меня нет ни трейсов, ни каких-либо метрик изнутри сервиса, сложно сказать, где именно могут быть узкие места. Но в целом, результаты хорошие.