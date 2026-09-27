DO $$
DECLARE
    v_region_id  UUID;
    v_server_id  UUID;
    v_ip         TEXT;
    v_host_num   INT;
BEGIN
    IF EXISTS (SELECT 1 FROM vpn_regions LIMIT 1) THEN
        RETURN;
    END IF;

    INSERT INTO vpn_regions (id, code, name)
    VALUES
        ('00000000-0000-0000-0000-000000000001', 'AF-NG', 'Africa — Nigeria'),
        ('00000000-0000-0000-0000-000000000002', 'EU-GB', 'Europe — United Kingdom'),
        ('00000000-0000-0000-0000-000000000003', 'NA-US', 'North America — United States');

    INSERT INTO vpn_servers (
        id, region_id, name, hostname, public_ip, country, city, provider,
        status, capacity, active_connections, wireguard_port, public_key
    ) VALUES
        (
            '10000000-0000-0000-0000-000000000001',
            '00000000-0000-0000-0000-000000000001',
            'NG-LAG-01', 'ng-lag-01.solidvpn.local', '127.0.0.1',
            'NG', 'Lagos', 'local-dev',
            'HEALTHY', 100, 0, 51820,
            'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY='
        ),
        (
            '10000000-0000-0000-0000-000000000002',
            '00000000-0000-0000-0000-000000000002',
            'GB-LON-01', 'gb-lon-01.solidvpn.local', '127.0.0.2',
            'GB', 'London', 'local-dev',
            'HEALTHY', 100, 0, 51820,
            'xJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY='
        ),
        (
            '10000000-0000-0000-0000-000000000003',
            '00000000-0000-0000-0000-000000000003',
            'US-NYC-01', 'us-nyc-01.solidvpn.local', '127.0.0.3',
            'US', 'New York', 'local-dev',
            'HEALTHY', 100, 0, 51820,
            'yJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY='
        );

    FOR v_server_id IN
        SELECT id FROM vpn_servers WHERE provider = 'local-dev'
    LOOP
        FOR v_host_num IN 2..101 LOOP
            v_ip := '10.8.' || ((v_host_num - 2) / 100) || '.' || ((v_host_num - 2) % 100 + 2);
            INSERT INTO ip_allocations (server_id, ip_address, allocated)
            VALUES (v_server_id, v_ip::INET, FALSE)
            ON CONFLICT DO NOTHING;
        END LOOP;
    END LOOP;
END $$;
