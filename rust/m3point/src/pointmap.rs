//! Murmur-hash chained map used by the Go point index.
//!
//! The Go map is a lock-free linked list per bucket. This port keeps the same
//! bucket function and chain walk, and guards each update with a mutex so the
//! concurrent tests observe the same results.

use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{Arc, Mutex};

use crate::murmur::murmur_hash_to_int;
use crate::point::Point;

struct Entry<V> {
    m_key: u32,
    key: Point,
    value: Arc<V>,
    next: Option<Box<Entry<V>>>,
}

pub struct PointHashMap<V> {
    entries: Mutex<Vec<Option<Box<Entry<V>>>>>,
    nb_elements: AtomicUsize,
    conflicts: AtomicUsize,
}

impl<V> PointHashMap<V> {
    pub fn new(init_size: usize) -> Self {
        let mut entries = Vec::with_capacity(init_size);
        for _ in 0..init_size {
            entries.push(None);
        }
        Self {
            entries: Mutex::new(entries),
            nb_elements: AtomicUsize::new(0),
            conflicts: AtomicUsize::new(0),
        }
    }

    pub fn init_size(&self) -> usize {
        self.entries.lock().expect("map lock").len()
    }

    pub fn size(&self) -> usize {
        self.nb_elements.load(Ordering::SeqCst)
    }

    pub fn conflicts(&self) -> usize {
        self.conflicts.load(Ordering::Relaxed)
    }

    pub fn has(&self, key: Point) -> bool {
        self.load(key).is_some()
    }

    pub fn load(&self, key: Point) -> Option<Arc<V>> {
        let m_key = key.murmur_hash();
        let entries = self.entries.lock().expect("map lock");
        let idx = bucket(m_key, entries.len());
        let mut entry = entries[idx].as_ref();
        while let Some(e) = entry {
            if e.key == key {
                return Some(Arc::clone(&e.value));
            }
            entry = e.next.as_ref();
        }
        None
    }

    /// Store `value`, returning the previous value when the key was already present.
    pub fn put(&self, key: Point, value: Arc<V>) -> Option<Arc<V>> {
        self.internal_put(key, value, true).0
    }

    /// Insert when absent. Returns the stored value and whether this call inserted it.
    pub fn load_or_store(&self, key: Point, value: Arc<V>) -> (Arc<V>, bool) {
        let (actual, inserted) = self.internal_put(key, value, false);
        (
            actual.expect("load_or_store always returns a value"),
            inserted,
        )
    }

    fn internal_put(
        &self,
        key: Point,
        value: Arc<V>,
        override_value: bool,
    ) -> (Option<Arc<V>>, bool) {
        let m_key = key.murmur_hash();
        let mut entries = self.entries.lock().expect("map lock");
        let idx = bucket(m_key, entries.len());
        let conflicts = &self.conflicts;
        let elements = &self.nb_elements;
        insert_at(
            &mut entries[idx],
            m_key,
            key,
            value,
            override_value,
            conflicts,
            elements,
        )
    }

    pub fn range<F>(&self, mut visit: F)
    where
        F: FnMut(&Point, &Arc<V>) -> bool,
    {
        let entries = self.entries.lock().expect("map lock");
        for bucket in entries.iter() {
            let mut entry = bucket.as_ref();
            while let Some(e) = entry {
                if visit(&e.key, &e.value) {
                    return;
                }
                entry = e.next.as_ref();
            }
        }
    }
}

fn bucket(m_key: u32, len: usize) -> usize {
    murmur_hash_to_int(m_key, len as i32) as usize
}

fn insert_at<V>(
    slot: &mut Option<Box<Entry<V>>>,
    m_key: u32,
    key: Point,
    value: Arc<V>,
    override_value: bool,
    conflicts: &AtomicUsize,
    elements: &AtomicUsize,
) -> (Option<Arc<V>>, bool) {
    if slot.is_none() {
        *slot = Some(Box::new(Entry {
            m_key,
            key,
            value: Arc::clone(&value),
            next: None,
        }));
        elements.fetch_add(1, Ordering::SeqCst);
        return if override_value {
            (None, true)
        } else {
            (Some(value), true)
        };
    }
    let entry = slot.as_mut().expect("slot just checked");
    if entry.m_key == m_key {
        if entry.key == key {
            if override_value {
                let old = Arc::clone(&entry.value);
                entry.value = value;
                return (Some(old), true);
            }
            return (Some(Arc::clone(&entry.value)), false);
        }
        conflicts.fetch_add(1, Ordering::Relaxed);
    }
    insert_at(
        &mut entry.next,
        m_key,
        key,
        value,
        override_value,
        conflicts,
        elements,
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::point::{CInt, Point};
    use std::collections::HashMap;
    use std::sync::atomic::{AtomicBool, AtomicI32};
    use std::thread;

    #[test]
    fn point_map_basic() {
        let m = PointHashMap::new(10);
        let p1 = Point([1, 2, 3]);
        let v23 = Arc::new(23i32);
        assert!(m.put(p1, Arc::clone(&v23)).is_none());
        assert_eq!(m.size(), 1);
        let got = m.load(Point([1, 2, 3])).unwrap();
        assert_eq!(*got, 23);
        assert!(Arc::ptr_eq(&got, &v23));
        assert!(m.load(Point([2, 2, 3])).is_none());
        assert_eq!(m.size(), 1);
        let v24 = Arc::new(24i32);
        assert!(m.put(Point([2, 2, 3]), Arc::clone(&v24)).is_none());
        assert_eq!(m.size(), 2);
        let v25 = Arc::new(25i32);
        let old = m.put(p1, Arc::clone(&v25)).unwrap();
        assert!(Arc::ptr_eq(&old, &v23));
        assert_eq!(*old, 23);
        assert_eq!(m.size(), 2);
        assert_eq!(*m.load(p1).unwrap(), 25);

        let p3 = Point([3, 3, 3]);
        let v26 = Arc::new(26i32);
        let (actual, inserted) = m.load_or_store(p3, Arc::clone(&v26));
        assert!(inserted);
        assert!(Arc::ptr_eq(&actual, &v26));
        assert_eq!(m.size(), 3);
        let (old, inserted) = m.load_or_store(Point([3, 3, 3]), Arc::new(27));
        assert!(!inserted);
        assert!(Arc::ptr_eq(&old, &v26));
        assert_eq!(*m.load(p3).unwrap(), 26);
        assert_eq!(m.size(), 3);
        assert!(m.has(p1));
    }

    #[test]
    fn point_map_conflicts_visit_each_point_once() {
        assert!(run_conflicts(25, 2, CInt::from(3)));
        assert!(run_conflicts(25 * 5, 8, CInt::from(5)));
        assert!(run_conflicts(5 * 3 * 3 * 3, 8, CInt::from(3)));
    }

    fn run_conflicts(hash_size: usize, nb_parallel: usize, rd_max: CInt) -> bool {
        let m = PointHashMap::<i64>::new(hash_size);
        let mut current = 0usize;
        for x in -rd_max..=rd_max {
            for y in -rd_max..=rd_max {
                for z in -rd_max..=rd_max {
                    assert_eq!(current, m.size());
                    let p = Point([x, y, z]);
                    assert!(m.put(p, Arc::new(p.distance_squared())).is_none());
                    current += 1;
                    assert_eq!(current, m.size());
                }
            }
        }
        let mut seen: HashMap<Point, Arc<AtomicI32>> = HashMap::new();
        m.range(|point, _| {
            assert!(!seen.contains_key(point), "visited {point} twice");
            seen.insert(*point, Arc::new(AtomicI32::new(1)));
            false
        });
        assert_eq!(m.size(), seen.len());

        let seen = Arc::new(Mutex::new(seen));
        let mut handles = Vec::new();
        let entries: Vec<Point> = {
            let guard = seen.lock().unwrap();
            guard.keys().copied().collect()
        };
        let chunk = (entries.len() / nb_parallel).max(1);
        for chunk in entries.chunks(chunk) {
            let chunk = chunk.to_vec();
            let seen = Arc::clone(&seen);
            handles.push(thread::spawn(move || {
                for point in chunk {
                    let map = seen.lock().unwrap();
                    let cell = map.get(&point).expect("point missing");
                    let prev = cell.fetch_add(1, Ordering::SeqCst);
                    assert_eq!(prev, 1, "{point}");
                }
            }));
        }
        for h in handles {
            h.join().unwrap();
        }
        true
    }

    #[test]
    fn point_map_concurrency() {
        run_concurrency(CInt::from(5), 4, 150, 0.5, false);
        run_concurrency(CInt::from(3), 2, 100, 0.2, false);
    }

    #[test]
    fn point_map_load_or_store() {
        run_concurrency(CInt::from(5), 4, 150, 0.5, true);
        run_concurrency(CInt::from(3), 2, 100, 0.2, true);
    }

    fn run_concurrency(
        rd_max: CInt,
        divider: usize,
        nb_routines: usize,
        hash_size_ratio: f64,
        load_and_store: bool,
    ) {
        let range_c = (rd_max + 1 + rd_max) as usize;
        let mut test_set = Vec::with_capacity(range_c * range_c * range_c);
        for x in -rd_max..=rd_max {
            for y in -rd_max..=rd_max {
                for z in -rd_max..=rd_max {
                    test_set.push(Point([x, y, z]));
                }
            }
        }
        let data_set_size = test_set.len();
        let nb_round = data_set_size / divider + divider - 1;
        assert!(nb_routines * nb_round > divider * data_set_size);
        assert!(nb_routines as f64 / divider as f64 > 16.0);

        let m = Arc::new(PointHashMap::<WasHere>::new(
            (data_set_size as f64 * hash_size_ratio) as usize,
        ));
        let failed = Arc::new(AtomicBool::new(false));
        let mut handles = Vec::new();
        for r in 0..nb_routines {
            let offset = (r % divider) * (data_set_size / divider);
            let m = Arc::clone(&m);
            let failed = Arc::clone(&failed);
            let test_set = test_set.clone();
            handles.push(thread::spawn(move || {
                for i in 0..nb_round {
                    if failed.load(Ordering::Relaxed) {
                        return;
                    }
                    let mut idx = offset + i;
                    if idx >= data_set_size {
                        idx = data_set_size - 1;
                    }
                    if load_and_store {
                        let (whl, _) = m.load_or_store(test_set[idx], Arc::new(WasHere::default()));
                        whl.add(r as i32);
                    } else {
                        m.put(test_set[idx], Arc::new(WasHere::default()));
                    }
                }
            }));
        }
        for h in handles {
            h.join().unwrap();
        }
        assert!(!failed.load(Ordering::Relaxed));
        assert_eq!(m.size(), data_set_size);

        if load_and_store {
            let mut handles = Vec::new();
            for r in 0..nb_routines {
                let offset = (r % divider) * (data_set_size / divider);
                let m = Arc::clone(&m);
                let failed = Arc::clone(&failed);
                let test_set = test_set.clone();
                handles.push(thread::spawn(move || {
                    for i in 0..nb_round {
                        if failed.load(Ordering::Relaxed) {
                            return;
                        }
                        let mut idx = offset + i;
                        if idx >= data_set_size {
                            idx = data_set_size - 1;
                        }
                        let whl = m.load(test_set[idx]).expect("missing point");
                        if !whl.has(r as i32) {
                            failed.store(true, Ordering::Relaxed);
                            return;
                        }
                    }
                }));
            }
            for h in handles {
                h.join().unwrap();
            }
            assert!(
                !failed.load(Ordering::Relaxed),
                "a routine did not find its id"
            );
        }
    }

    #[derive(Default)]
    struct WasHere {
        ids: Mutex<Vec<i32>>,
    }

    impl WasHere {
        fn add(&self, id: i32) {
            let mut ids = self.ids.lock().unwrap();
            if !ids.contains(&id) {
                ids.push(id);
            }
        }
        fn has(&self, id: i32) -> bool {
            self.ids.lock().unwrap().contains(&id)
        }
    }
}
