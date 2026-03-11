INSERT INTO incidents (id, service, environment, severity, status, message, trace_id, metadata, created_at, resolved_at)
SELECT
    gen_random_uuid(),
    'go-service-template-pro',
    'local',
    data.severity,
    data.status,
    data.message,
    data.trace_id,
    data.metadata::jsonb,
    NOW() - (data.offset_minutes || ' minutes')::interval,
    CASE WHEN data.status = 'resolved' THEN NOW() - ((data.offset_minutes - 10) || ' minutes')::interval ELSE NULL END
FROM (
    VALUES
        ('critical', 'open', 'Payment gateway timeout spike', 'd4aa741b57f5f35c3d12cced5e001111', '{"region":"eu-west-1","tenant":"enterprise"}', 8),
        ('high', 'investigating', 'Increased p95 latency above SLO', 'd4aa741b57f5f35c3d12cced5e002222', '{"slo":"p95<120ms"}', 20),
        ('medium', 'resolved', 'Cache miss ratio exceeded threshold', 'd4aa741b57f5f35c3d12cced5e003333', '{"cache":"redis"}', 55),
        ('low', 'resolved', 'Scheduled batch retried once', 'd4aa741b57f5f35c3d12cced5e004444', '{"job":"hourly-reconciliation"}', 95)
) AS data(severity, status, message, trace_id, metadata, offset_minutes)
WHERE NOT EXISTS (SELECT 1 FROM incidents LIMIT 1);
