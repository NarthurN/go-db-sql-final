package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")// настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	lastInsertedId, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, lastInsertedId)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	insertedParcel, err := store.Get(lastInsertedId)
	require.NoError(t, err)
	assert.NotEmpty(t, insertedParcel.Number)
	parcel.Number = lastInsertedId
	assert.Equal(t, parcel, insertedParcel)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(lastInsertedId)
	require.NoError(t, err)

	_, err = store.Get(lastInsertedId)
	require.ErrorIs(t, err, sql.ErrNoRows)
}


// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")// настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	lastInsertedID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, lastInsertedID)
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(lastInsertedID, newAddress)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	insertedParcel, err := store.Get(lastInsertedID)
	require.NoError(t, err)
	require.Equal(t, newAddress, insertedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")// настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	lastInsertedID, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, lastInsertedID)
	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	newStatus := "new test status"
	err = store.SetStatus(lastInsertedID, newStatus)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	insertedParcel, err := store.Get(lastInsertedID)
	require.NoError(t, err)
	require.Equal(t, newStatus, insertedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")// настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		p, ok := parcelMap[parcel.Number] 
		assert.True(t, ok)
		assert.Equal(t, p, parcel)
	}
}
