'use client';

import { useEffect } from 'react';
import { initializeFrameworks } from '@/shared/frameworks/init';

export function ClientInitializer() {
    useEffect(() => {
        console.log('[ClientInitializer] Initializing framework registry...');
        initializeFrameworks();
    }, []);

    return null;
}
